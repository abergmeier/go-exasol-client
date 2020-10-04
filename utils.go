/*
    AUTHOR

	Grant Street Group <developers@grantstreet.com>

	COPYRIGHT AND LICENSE

	This software is Copyright (c) 2019 by Grant Street Group.
	This is free software, licensed under:
	    MIT License
*/

/*--- Various utility routines ---*/

package exasol

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/davecgh/go-spew/spew"
)

var keywordLock sync.RWMutex
var keywords map[string]bool

/*--- Public Interface ---*/

func (c *Conn) QuoteIdent(ident string) string {
	if keywords == nil {
		keywordLock.Lock()
		if keywords == nil {
			keywords = map[string]bool{}
			sql := "SELECT LOWER(keyword) FROM sys.exa_sql_keywords WHERE reserved"
			kwRes, _ := c.FetchChan(sql)
			for col := range kwRes {
				keywords[col[0].(string)] = true
			}
		}
		keywordLock.Unlock()
	}
	_, isKeyword := keywords[strings.ToLower(ident)]
	if isKeyword {
		return fmt.Sprintf(`[%s]`, strings.ToLower(ident))
	} else if regexp.MustCompile(`^[^A-Za-z]`).MatchString(ident) {
		return fmt.Sprintf(`[%s]`, strings.ToUpper(ident))
	}
	return ident
}

func QuoteStr(str string) string {
	return regexp.MustCompile("'").ReplaceAllString(str, "''")
}

func Transpose(matrix [][]interface{}) [][]interface{} {
	numRows := len(matrix)
	numCols := len(matrix[0])
	ret := make([][]interface{}, numCols)

	for x, _ := range ret {
		ret[x] = make([]interface{}, numRows)
	}
	for y, s := range matrix {
		for x, e := range s {
			ret[x][y] = e
		}
	}
	return ret
}

/*--- Private Routines ---*/

func (c *Conn) error(args ...interface{}) {
	if c.Conf.SuppressError == false {
		log.Error(args...)
	}
}

func transposeToChan(ch chan<- []interface{}, matrix []interface{}) {
	// matrix is columnar ... this transposes it to rowular
	for row := range matrix[0].([]interface{}) {
		ret := make([]interface{}, len(matrix))
		for col := range matrix {
			ret[col] = matrix[col].([]interface{})[row]
		}
		ch <- ret
	}
}

// For debugging
func warnJson(msg interface{}) {
	json, _ := json.Marshal(msg)
	log.Errorf("WARNING: %s", json)
}

func dieJson(msg interface{}) {
	json, _ := json.Marshal(msg)
	log.Panicf("DIEING: %s", json)
}

func dump(i interface{}) {
	spew.Dump(i)
}