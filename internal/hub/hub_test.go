package hub

import "testing"

// TestBroadcastSkipsSenderAndDelivers verifica que el broadcast funcione correctamente:
// - El dispositivo emisor NO debe recibir su propio mensaje
// - Los demás dispositivos del mismo usuario SÍ deben recibir el mensaje
func TestBroadcastSkipsSenderAndDelivers(t *testing.T) {
	// Crea un Hub con buffer de 1 para los canales
	h := New(1)

	// Conecta el dispositivo A del usuario1
	chA, leaveA := h.Join("user1", "A")
	defer leaveA() // Asegura la limpieza al final del test

	// Conecta el dispositivo B del mismo usuario1
	chB, leaveB := h.Join("user1", "B")
	defer leaveB() // Asegura la limpieza al final del test

	// El dispositivo A envía un mensaje broadcast
	h.Broadcast("user1", "A", []byte("hola"))

	// Verifica que A (emisor) NO recibe su propio mensaje
	select {
	case <-chA:
		t.Fatal("el emisor no debería recibir su propio mensaje")
	default:
		// Correcto: no hay mensaje en el canal de A
	}

	// Verifica que B SÍ recibe el mensaje enviado por A
	select {
	case got := <-chB:
		if string(got) != "hola" {
			t.Fatalf("esperaba 'hola', obtuve %q", got)
		}
		// Correcto: B recibió el mensaje esperado
	default:
		t.Fatal("B no recibió el mensaje")
	}
}
