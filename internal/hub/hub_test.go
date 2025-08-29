package hub

import (
	"testing"
	"time"
)

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

// TestBroadcastDoesNotBlockWhenReceiverFull verifica que el broadcast no se bloquee
// cuando un canal receptor está lleno. En su lugar, debería descartar mensajes
// usando select con default para evitar deadlocks.
func TestBroadcastDoesNotBlockWhenReceiverFull(t *testing.T) {
	// Crea Hub con buffer de 1 (se llenará rápidamente)
	h := New(1)

	// Conecta dispositivo B del usuario1
	chB, leaveB := h.Join("user1", "B")
	defer leaveB() // Limpieza al final del test

	// Primer broadcast llena el buffer del canal de B
	h.Broadcast("user1", "A", []byte("x"))

	// Canal para verificar que el broadcast no se bloquea
	done := make(chan struct{})

	// Ejecuta segundo broadcast en goroutine separada
	go func() {
		// Este broadcast debería descartar el mensaje (canal lleno) sin bloquear
		h.Broadcast("user1", "A", []byte("y"))
		close(done) // Señala que terminó sin bloqueo
	}()

	// Verifica que el broadcast termine rápidamente (no se bloquea)
	select {
	case <-done:
		// OK: el broadcast no se bloqueó
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Broadcast se bloqueo con canal lleno")
	}

	// Verifica que B recibió el primer mensaje
	var got []byte
	select {
	case got = <-chB:
		if string(got) != "x" {
			t.Fatalf("quería 'x', llegó %q", got)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("B no recibió el primer mensaje")
	}

	// Verifica que el segundo mensaje fue descartado (canal estaba lleno)
	select {
	case <-chB:
		t.Fatal("debería haberse descartado el segundo mensaje")
	default:
		// OK: no hay más mensajes, el segundo fue descartado correctamente
	}
}

// TestLeaveRemovesAndClosesChannel verifica que la función leave() funcione correctamente:
// - Debe cerrar el canal del dispositivo que se desconecta
// - Debe remover el dispositivo del room del usuario
// - No debe afectar a otros dispositivos del mismo usuario
// - Debe permitir broadcasts posteriores sin errores
func TestLeaveRemovesAndClosesChannel(t *testing.T) {
	// Crea Hub con buffer de 1
	h := New(1)

	// Conecta dispositivo A del usuario1
	chA, leaveA := h.Join("user1", "A")
	defer leaveA() // Limpieza al final del test

	// Conecta dispositivo B del mismo usuario1
	chB, leaveB := h.Join("user1", "B")

	// Dispositivo B se desconecta explícitamente
	leaveB()

	// Verifica que el canal de B se cerró correctamente
	select {
	case _, ok := <-chB:
		if ok {
			t.Fatal("chB debería estar cerrado tras leaveB()")
		}
		// OK: canal cerrado (ok == false)
	default:
		t.Fatal("esperaba canal cerrado inmediatamente después de leaveB()")
	}

	// Envía broadcast para verificar que el sistema sigue funcionando
	// aunque B ya no esté conectado
	h.Broadcast("user1", "A", []byte("hola"))

	// Verifica que no hay errores/pánico tras broadcast con dispositivo desconectado
	select {
	case <-chA:
		// OK: A podría recibir si hubiera otro emisor;
		// aquí solo verificamos que no hay pánico.
	default:
		// También OK: A es el emisor, no debería recibir su propio mensaje
	}
}

func TestRoomsAreIsolated(t *testing.T) {
	h := New(1)

	chU1, leaveU1 := h.Join("user1", "A")
	defer leaveU1()

	chU2, leaveU2 := h.Join("user2", "X")
	defer leaveU2()

	h.Broadcast("user1", "A", []byte("hola"))

	select {
	case <-chU2:
		t.Fatal("user2 no debería recibir mensajes de user1")
	default:
	}

	select {
	case <-chU1:
		t.Fatal("el emisor no debería recibir su propio mensaje")
	default:
	}
}
