package util

import (
	"context"
	"fmt"
	"github.com/neflyte/configmap"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
	"github.com/thoas/go-funk"
	"net/url"
	"regexp"
	"strings"
)

const (
	PGConnHost    = "host"
	PGConnPort    = "port"
	PGConnDBName  = "dbname"
	PGConnUser    = "user"
	PGConnSSLMode = "sslmode"
)

var (
	StartsWithNumberRE   = regexp.MustCompile(`^[0-9].*$`)
	pgConnectionStringRE = regexp.MustCompile(`([\w]+)[ ]?=[ ]?([\w]+)`)
	urlStringRE          = regexp.MustCompile(`^(postgres[q]?[l]?:\/\/)?([^:\/\s]+)(:([^\/]*))?(\/\w+\.)*([^#?\s]+)(\?([^#]*))?(#(.*))?$`)
	pgConnectionFields   = []interface{}{PGConnHost, PGConnPort, PGConnDBName, PGConnUser, PGConnSSLMode}
)

func StringCoalesce(args ...string) string {
	for _, str := range args {
		if strings.TrimSpace(str) != "" {
			return str
		}
	}
	return ""
}

func StartsWithNumber(str string) bool {
	return StartsWithNumberRE.MatchString(str)
}

func MapContainsKeys(haystack configmap.ConfigMap, needles []interface{}) bool {
	return funk.Every(funk.Keys(haystack), needles...)
}

func ConnectionStringWithSecret(connURL *url.URL, secret model.Secret) string {
	log := logger.Root().
		WithField("function", "ConnectionStringWithSecret")
	secretValue, err := GetSecret(context.Background(), secret)
	if err != nil {
		log.Errorf("error getting secret value: %s; returning connection string as-is", err)
		log.Tracef("returning %s", connURL.String())
		return connURL.String()
	}
	connURL.User = url.UserPassword(connURL.User.Username(), secretValue)
	return connURL.String()
}

func ResolveConnectionString(connStr string, secret *model.Secret) string {
	var connURL *url.URL
	var err error

	log := logger.Root().
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
		if MapContainsKeys(connMap, pgConnectionFields) {
			log.Trace("connMap has enough keys")
			// Build the url.URL host string
			urlHost := fmt.Sprintf("%s:%s", connMap.GetString(PGConnHost), connMap.GetString(PGConnPort))
			// Build the query string
			urlQuery := fmt.Sprintf("%s=%s", PGConnSSLMode, connMap.GetString(PGConnSSLMode))
			// Create a new url.URL
			connURL = &url.URL{
				Scheme:   "postgres",
				User:     url.User(connMap.GetString(PGConnUser)),
				Host:     urlHost,
				Path:     connMap.GetString(PGConnDBName),
				RawQuery: urlQuery,
			}
		} else {
			log.Debugf("map didn't contain all keys; map: %#v, keys: %#v", connMap, pgConnectionFields)
		}
	}
	// Handle secret
	if SecretIsDefined(*secret) {
		return ConnectionStringWithSecret(connURL, *secret)
	}
	return connURL.String()
}
