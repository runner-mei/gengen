package types

import (
	"github.com/grsmv/inflect"
)

// func CamelCase(name string) string {
// 	if "db2" == name {
// 		return "DB2"
// 	}
// 	return inflect.Camelize(name)
// }

func Underscore(name string) string {
	switch name {
	case "DB2":
		return "db2"
	case "IIS":
		return "iis"
	case "TCP":
		return "tcp"
	case "HTTP":
		return "http"
	case "FTP":
		return "ftp"
	case "SMTP":
		return "smtp"
	case "POP3":
		return "pop3"
	case "IMAP":
		return "imap"
	case "DHCP":
		return "dhcp"
	case "DNS":
		return "dns"
	case "LDAP":
		return "ldap"
	case "EPON":
		return "epon"
	case "MplsPE":
		return "mpls_pe"
	case "MplsCE":
		return "mpls_ce"
	}
	return inflect.Underscore(name)
}

// func Pluralize(name string) string {
// 	if "db2" == name {
// 		return "db2"
// 	}

// 	return inflect.Pluralize(name)
// }

func Tableize(className string) string {
	switch className {
	case "DB2":
		return "db2"
	case "IIS":
		return "iis"
	case "TCP":
		return "tcp"
	case "HTTP":
		return "http"
	case "FTP":
		return "ftp"
	case "SMTP":
		return "smtp"
	case "POP3":
		return "pop3"
	case "IMAP":
		return "imap"
	case "DHCP":
		return "dhcp"
	case "DNS":
		return "dns"
	case "LDAP":
		return "ldap"
	case "EPON":
		return "epon"
	case "MplsPE":
		return "mpls_pe"
	case "MplsCE":
		return "mpls_ce"
	}
	return inflect.Pluralize(inflect.Underscore(className))
}

func Singularize(word string) string {
	if "windows" == word {
		return "windows"
	}
	if "iis" == word {
		return "iis"
	}
	return inflect.Singularize(word)
}
