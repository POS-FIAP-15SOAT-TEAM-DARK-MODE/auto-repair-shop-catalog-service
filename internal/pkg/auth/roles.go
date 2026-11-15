package auth

// Role identifies a user role carried in the JWT "roles" claim.
type Role string

const (
	ADMIN     Role = "ADMIN"
	ATTENDANT Role = "ATTENDANT"
	MECHANIC  Role = "MECHANIC"
	CUSTOMER  Role = "CUSTOMER"
)

var (
	AttendantRoles            = []Role{ATTENDANT, ADMIN}
	AttendantAndMechanicRoles = []Role{ATTENDANT, MECHANIC, ADMIN}
)
