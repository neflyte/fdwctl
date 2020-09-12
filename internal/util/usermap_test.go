package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUnit_PasswordRE(t *testing.T) {
	query := `CREATE USER MAPPING FOR foouser SERVER barserver OPTIONS (user 'baz', password 'zot')`
	require.True(t, PasswordRE.MatchString(query))
	noPassword := `ALTER USER MAPPING FOR foouser SERVER barserver OPTIONS ( SET user 'baz' )` //nolint:gosec
	require.False(t, PasswordRE.MatchString(noPassword))
}

func TestUnit_SanitizePasswordInSQL_Nominal(t *testing.T) {
	query := `CREATE USER MAPPING FOR foouser SERVER barserver OPTIONS (user 'baz', password 'zot')`
	expected := `CREATE USER MAPPING FOR foouser SERVER barserver OPTIONS (user 'baz', password '...')`
	actual := SanitizePasswordInSQL(query)
	require.Equal(t, expected, actual)
}

func TestUnit_SanitizePasswordInSQL_NoPassword(t *testing.T) {
	query := `ALTER USER MAPPING FOR foouser SERVER barserver OPTIONS ( SET user 'baz' )`
	expected := `ALTER USER MAPPING FOR foouser SERVER barserver OPTIONS ( SET user 'baz' )`
	actual := SanitizePasswordInSQL(query)
	require.Equal(t, expected, actual)
}
