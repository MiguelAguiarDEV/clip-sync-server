package hub

import (
	"sync"
)

// sub representa una suscripción de dispositivo a un room de usuario.
// Cada dispositivo conectado tiene su propia instancia de sub.
type sub struct {
	deviceID string      // Identificador único del dispositivo
	ch       chan []byte // Canal con buffer para recibir mensajes broadcast
}

// Hub gestiona múltiples rooms de usuarios, donde cada room contiene
// dispositivos conectados que pueden intercambiar mensajes.
// Es thread-safe para uso concurrente.
type Hub struct {
	mu     sync.RWMutex               // Mutex para acceso concurrente seguro
	rooms  map[string]map[string]*sub // userID -> deviceID -> suscripción
	bufCap int                        // Capacidad del buffer para canales de dispositivos
}

// New crea una nueva instancia de Hub con capacidad de buffer especificada.
// El Hub gestiona rooms de usuarios donde cada room contiene dispositivos conectados.
// Si bufCap <= 0, usa un valor por defecto de 32 para el buffer de cada canal.
// Retorna un puntero al Hub inicializado.
func New(bufCap int) *Hub {
	if bufCap <= 0 {
		bufCap = 32
	}
	return &Hub{
		rooms:  make(map[string]map[string]*sub),
		bufCap: bufCap,
	}
}

// Join registra un dispositivo en el room del usuario especificado.
// Crea un canal con buffer para que el dispositivo reciba mensajes.
// Si el room del usuario no existe, lo crea automáticamente.
// Retorna:
//   - Un canal de solo lectura para recibir mensajes broadcast
//   - Una función leave() para desconectar el dispositivo y limpiar recursos
//
// La función leave() es thread-safe y cierra el canal automáticamente.
func (h *Hub) Join(userID, deviceID string) (<-chan []byte, func()) {
	ch := make(chan []byte, h.bufCap) // canal con buffer por dispositivo
	s := &sub{deviceID: deviceID, ch: ch}

	h.mu.Lock()
	if _, ok := h.rooms[userID]; !ok {
		h.rooms[userID] = make(map[string]*sub)
	}
	h.rooms[userID][deviceID] = s
	h.mu.Unlock()

	leave := func() {
		h.mu.Lock()
		if m, ok := h.rooms[userID]; ok {
			// elimina solo si es el mismo puntero (evita carreras con re-join)
			if m[deviceID] == s {
				delete(m, deviceID)
			}
			if len(m) == 0 {
				delete(h.rooms, userID)
			}
		}
		h.mu.Unlock()
		close(ch)
	}
	return ch, leave
}

// Broadcast envía un mensaje a todos los dispositivos de un usuario excepto al remitente.
// Parámetros:
//   - userID: identificador del usuario cuyo room recibirá el mensaje
//   - fromDevice: identificador del dispositivo que envía (se excluye del broadcast)
//   - payload: datos a enviar a los dispositivos
func (h *Hub) Broadcast(userID, fromDevice string, payload []byte) {
	// Adquiere lock de lectura para acceso concurrente seguro
	// (Cierra el candado para que nadie pueda escribir en el room mientras estamos mirando dentro)
	h.mu.RLock()

	// Busca el room del usuario (Si no encuentra un room asignado al userId hace return sino almacena todos los dispositivos)
	m, ok := h.rooms[userID]
	if !ok {
		// Si el usuario no tiene room, sale sin hacer nada
		h.mu.RUnlock()
		return
	}

	// Recolecta todos los dispositivos del usuario excepto el remitente
	subs := make([]*sub, 0, len(m))
	for id, s := range m {
		if id != fromDevice {
			subs = append(subs, s)
		}
	}
	h.mu.RUnlock()

	for _, s := range subs {
		select {
		case s.ch <- payload: // intenta enviar
		default: // si la cola está llena, lo saltamos (no bloquea)
		}
	}
}
