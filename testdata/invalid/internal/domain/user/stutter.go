package user

type UserService struct{} // violation: stutters
type Service struct{}     // OK
type PowerUser struct{}   // OK: "User" is not prefix
