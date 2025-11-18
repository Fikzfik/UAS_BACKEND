package models

type Permission struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Resource    string `json:"resource"`
    Action      string `json:"action"`
    Description string `json:"description"`
}

type RolePermission struct {
    RoleID       string `json:"role_id"`
    PermissionID string `json:"permission_id"`
}