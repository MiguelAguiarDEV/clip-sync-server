package types

const MaxInlineBytes = 64 * 1024 // 64KB

type Hello struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
}

type Clip struct {
	MsgID     string `json:"msg_id"`
	Mime      string `json:"mime"`
	Size      int    `json:"size"`
	Data      []byte `json:"data,omitempty"`       // si es pequeño (≤ MaxInlineBytes)
	UploadURL string `json:"upload_url,omitempty"` // si es grande
	From      string `json:"from"`                 // lo rellenará el server
}

type Envelope struct {
	Type  string `json:"type"` // "hello" | "clip"
	Hello *Hello `json:"hello,omitempty"`
	Clip  *Clip  `json:"clip,omitempty"`
}
