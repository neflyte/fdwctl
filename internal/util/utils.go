/*
Package util contains various utility methods and constants
*/
package util

import (
	"context"
	"fmt"
	"github.com/neflyte/configmap"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
	"net/url"
	"regexp"
	"strings"
)

const (
	// pgConnHost is the name of the Host connection string key
	pgConnHost = "host"
	// pgConnPort is the name of the Port connection string key
	pgConnPort = "port"
	// pgConnDBName is the name of the Database connection string key
	pgConnDBName = "dbname"
	// pgConnUser is the name of the User connection string key
	pgConnUser = "user"
	// pgConnSSLMode is the name of the SSLMode connection string key
	pgConnSSLMode = "sslmode"
)

var (
	// startsWithNumberRE is a regular expression to test if a string begins with a number
	startsWithNumberRE = regexp.MustCompile(`^[0-9].*$`)
	// pgConnectionStringRE is a regular expression that matches a PG-style connection string
	pgConnectionStringRE = regexp.MustCompile(`([\w]+)[ ]?=[ ]?([\w]+)`)
	// urlStringRE is a regular expression that matches a URL-style connection string
	urlStringRE = regexp.MustCompile(`^(postgres[q]?[l]?://)?([^:/\s]+)(:([^/]*))?(/\w+\.)*([^#?\s]+)(\?([^#]*))?(#(.*))?$`)
	// pgConnectionFields is a list of connection string fields that must exist for the string to be considered valid
	pgConnectionFields = []string{pgConnHost, pgConnPort, pgConnDBName, pgConnUser, pgConnSSLMode}
)

// StringCoalesce returns the first string in the supplied arguments that is non-empty when trimmed. If there
// is no such string, the empty string is returned.
func StringCoalesce(args ...string) string {
	for _, str := range args {
		if strings.TrimSpace(str) != "" {
			return str
		}
	}
	return ""
}

// StartsWithNumber determines if a string starts with a number
func StartsWithNumber(str string) bool {
	return startsWithNumberRE.MatchString(str)
}

// mapContainsKeys determines if the supplied map (haystack) contains all the supplied keys (needles)
func mapContainsKeys(haystack configmap.ConfigMap, needles ...string) bool {
	needlecount := 0
	for _, needle := range needles {
		if haystack.Has(needle) {
			needlecount++
		}
	}
	return needlecount == len(needles)
}

// connectionStringWithSecret returns a URL populated with a credential obtained using the supplied secret configuration
func connectionStringWithSecret(connURL *url.URL, secret model.Secret) string {
	log := logger.Log().
		WithField("function", "connectionStringWithSecret")
	secretValue, err := GetSecret(context.Background(), secret)
	if err != nil {
		log.Errorf("error getting secret value: %s; returning connection string as-is", err)
		log.Tracef("returning %s", connURL.String())
		return connURL.String()
	}
	connURL.User = url.UserPassword(connURL.User.Username(), secretValue)
	return connURL.String()
}

// ResolveConnectionString returns a connection string populated with a credential obtained using the supplied secret configuration
func ResolveConnectionString(connStr string, secret *model.Secret) string {
	var connURL *url.URL
	var err error

	log := logger.Log().
		WithField("function", "ResolveConnectionString")
	if connStr == "" {
		return ""
	}
	/*
		There are two kinds of connection string:
		1. URL-style (e.g. RFC-3986: postgres://user:password@host:port/db?options...)
		2. PG-style (e.g. host=xxx port=yyy db=zzz...)
	*/
	// Do we have an URL-style string?
	if urlStringRE.MatchString(connStr) {
		connURL, err = url.Parse(connStr)
		if err != nil {
			log.Errorf("error parsing connection string as URL: %s", err)
		} else {
			log.Trace("got an URL-style string")
		}
	}
	// Do we have a PG-style string?
	if connURL == nil && pgConnectionStringRE.MatchString(connStr) {
		log.Trace("found PG-style string")
		connMap := configmap.New()
		matches := pgConnectionStringRE.FindAllStringSubmatch(connStr, -1)
		for _, match := range matches {
			log.Tracef("match: %#v", match)
			connMap.Set(match[1], match[2])
		}
		// Now we can try to construct an URL-style string. First we see if there
		// is enough information to do so
		if mapContainsKeys(connMap, pgConnectionFields...) {
			log.Trace("connMap has enough keys")
			// Build the url.URL host string
			urlHost := fmt.Sprintf("%s:%s", connMap.GetString(pgConnHost), connMap.GetString(pgConnPort))
			// Build the query string
			urlQuery := fmt.Sprintf("%s=%s", pgConnSSLMode, connMap.GetString(pgConnSSLMode))
			// Create a new url.URL
			connURL = &url.URL{
				Scheme:   "postgres",
				User:     url.User(connMap.GetString(pgConnUser)),
				Host:     urlHost,
				Path:     connMap.GetString(pgConnDBName),
				RawQuery: urlQuery,
			}
		} else {
			log.Debugf("map didn't contain all keys; map: %#v, keys: %#v", connMap, pgConnectionFields)
		}
	}
	// Handle secret
	if SecretIsDefined(*secret) {
		return connectionStringWithSecret(connURL, *secret)
	}
	return connURL.String()
}
