package dto

import "time"

// MessageMediaMetadataRequest merepresentasikan metadata media untuk request.
type MessageMediaMetadataRequest struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

// MessageReactionRequest merepresentasikan reaksi untuk request.
type MessageReactionRequest struct {
	Emoji  string `json:"emoji" validate:"required"`
	UserID string `json:"userId" validate:"required"`
}

// MessageCreateRequest merepresentasikan data yang diterima saat membuat pesan baru.
type MessageCreateRequest struct {
	ChannelID     string                       `json:"channelId" validate:"required"`
	UserID        string                       `json:"userId" validate:"required"`
	Content       string                       `json:"content,omitempty"`
	MessageType   string                       `json:"messageType" validate:"required" enums:"text,image,file,voice"`
	MediaPath     string                       `json:"mediaPath,omitempty"`
	MediaMetadata *MessageMediaMetadataRequest `json:"mediaMetadata,omitempty"`
}

// MessageUpdateRequest merepresentasikan data yang diterima saat memperbarui pesan.
type MessageUpdateRequest struct {
	Content       string                       `json:"content,omitempty"`
	MessageType   string                       `json:"messageType,omitempty" enums:"text,image,file,voice"`
	MediaPath     string                       `json:"mediaPath,omitempty"`
	MediaMetadata *MessageMediaMetadataRequest `json:"mediaMetadata,omitempty"`
	IsPinned      *bool                        `json:"isPinned,omitempty"`
}

// MessageReactionAddRequest merepresentasikan data untuk menambahkan reaksi.
type MessageReactionAddRequest struct {
	Emoji  string `json:"emoji" validate:"required"`
	UserID string `json:"userId" validate:"required"`
}

// MessageReactionRemoveRequest merepresentasikan data untuk menghapus reaksi.
type MessageReactionRemoveRequest struct {
	Emoji  string `json:"emoji" validate:"required"`
	UserID string `json:"userId" validate:"required"`
}

// MessageResponse merepresentasikan data pesan yang dikirimkan sebagai respons API.
type MessageResponse struct {
	ID            string                       `json:"id"`
	ChannelID     string                       `json:"channelId"`
	UserID        string                       `json:"userId"`
	Content       string                       `json:"content"`
	MessageType   string                       `json:"messageType"`
	MediaPath     string                       `json:"mediaPath,omitempty"`
	MediaMetadata *MessageMediaMetadataRequest `json:"mediaMetadata,omitempty"`
	CreatedAt     time.Time                    `json:"createdAt"`
	UpdatedAt     *time.Time                   `json:"updatedAt,omitempty"`
	IsPinned      bool                         `json:"isPinned"`
	Reactions     []MessageReactionRequest     `json:"reactions"`
}
