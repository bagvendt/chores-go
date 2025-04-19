package contextkeys

// ContextKey is a type for request context keys.
type ContextKey string

// UserContextKey is the key used to store the authenticated user in the request context.
const UserContextKey = ContextKey("user")
