// Command devtoken prints a development JWT signed with JWT_SECRET, for local calls to /v1.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
)

const defaultUserID = "33333333-3333-4333-8333-000000000001" // seeded attendant

var errUnknownRole = errors.New("unknown role: use ADMIN, ATTENDANT, MECHANIC or CUSTOMER")

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("devtoken", flag.ContinueOnError)
	role := fs.String("role", string(auth.ATTENDANT), "role claim: ADMIN, ATTENDANT, MECHANIC or CUSTOMER")
	userID := fs.String("user-id", defaultUserID, "user_id claim")
	ttl := fs.Duration("ttl", time.Hour, "token lifetime")
	if err := fs.Parse(args); err != nil {
		return err
	}

	r := auth.Role(*role)
	switch r {
	case auth.ADMIN, auth.ATTENDANT, auth.MECHANIC, auth.CUSTOMER:
	default:
		return errUnknownRole
	}

	token, err := auth.GenerateToken(*userID, []auth.Role{r}, time.Now().Add(*ttl))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, token)
	return err
}
