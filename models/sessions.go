package models

import "time"

type Session struct {
    ID              uint64    `json:"id"`
    Code            string    `json:"code"`
    Name            string    `json:"name"`
    OwnerID         uint64    `json:"owner_id,omitempty"`
    MaxCollaborators int      `json:"max_collaborators"`
    Status          string    `json:"status"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
