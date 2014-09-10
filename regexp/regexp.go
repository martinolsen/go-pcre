// Copyright 2014 Martin Olsen, and others. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This package is intended as a PCRE compatible drop-in replacement
// the standard regexp package.
package regexp

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"runtime"

	"github.com/martinolsen/go-pcre"
)

func MatchString(pattern string, s string) (bool, error) {
	if re, err := Compile(pattern); err != nil {
		return false, err
	} else if matched := re.MatchString(s); !matched {
		return false, nil
	} else {
		return true, nil
	}
}

type Regexp struct {
	expr string
	pcre *pcre.PCRE
}

func Compile(expr string) (*Regexp, error) {
	// TODO - pcre.Study

	if re, err := pcre.Compile(expr, pcre.Utf8|pcre.Dupnames, nil); err != nil {
		return nil, err
	} else {
		runtime.SetFinalizer(re, re.Free)
		return &Regexp{expr: expr, pcre: re}, nil
	}
}

func CompilePOSIX(expr string) (*Regexp, error) {
	return nil, fmt.Errorf("TODO - CompilePOSIX")
}

func Match(pattern string, b []byte) (matched bool, err error) {
	return false, fmt.Errorf("TODO - Match")
}

func MustCompile(str string) *Regexp {
	if re, err := Compile(str); err != nil {
		panic(err)
	} else {
		return re
	}
}

func MustCompilePOSIX(str string) *Regexp {
	if re, err := CompilePOSIX(str); err != nil {
		panic(err)
	} else {
		return re
	}
}

func (re *Regexp) Match(b []byte) bool {
	return re.MatchString(string(b))
}

func (re *Regexp) MatchString(s string) bool {
	if err := re.pcre.Exec(nil, s, 0, 0, nil); err == pcre.ErrNomatch {
		return false
	} else if err < 0 {
		panic("dont know what to do")
	}

	return true
}

func MatchReader(pattern string, r io.RuneReader) (matched bool, err error) {
	return false, fmt.Errorf("TODO - MatchReader")
}

func (re *Regexp) Find(b []byte) []byte {
	if is := re.FindIndex(b); len(is) != 2 {
		return nil
	} else {
		return b[is[0]:is[1]]
	}
}

func (re *Regexp) FindString(s string) string { return string(re.Find([]byte(s))) }

func (re *Regexp) FindAll(b []byte, n int) [][]byte {
	var c [][]byte
	for _, loc := range re.FindAllIndex(b, n) {
		c = append(c, b[loc[0]:loc[1]])
	}
	return c
}

func (re *Regexp) FindAllIndex(b []byte, n int) [][]int {
	var locs [][]int
	var options pcre.Option
	for start := 0; start <= len(b) && n != 0; n-- {
		ovector := make([]int, 3)
		if e := re.pcre.Exec(nil, string(b), start, options, ovector); e == pcre.ErrNomatch {
			break
		} else if e < 0 {
			log.Panicf("while mathcing %q[%d:]: e: %d", b, start, e)
		} else {
			locs = append(locs, []int{ovector[0], ovector[1]})
		}
		options |= pcre.NotemptyAtstart
		start = ovector[1]
	}

	return locs
}

func (re *Regexp) FindAllString(s string, n int) []string {
	var ss []string
	for _, s := range re.FindAll([]byte(s), n) {
		ss = append(ss, string(s))
	}
	return ss
}

func (re *Regexp) FindAllStringIndex(s string, n int) [][]int { return re.FindAllIndex([]byte(s), n) }

func (re *Regexp) FindIndex(b []byte) []int {
	ovector := make([]int, 3)

	if e := re.pcre.Exec(nil, string(b), 0, 0, ovector); e < pcre.ErrNomatch {
		log.Panicf("e: %s", e)
	} else if e >= 0 {
		return []int{ovector[0], ovector[1]}
	}

	return nil
}

func (re *Regexp) FindStringIndex(s string) []int { return re.FindIndex([]byte(s)) }

func (re *Regexp) FindReaderIndex(r io.RuneReader) (loc []int) {
	data, err := readAllRunes(r)
	if err != nil {
		log.Panicf("readAllRunes: %s", err)
	}

	return re.FindIndex(data)
}

func (re *Regexp) FindAllStringSubmatch(s string, n int) [][]string {
	var rs [][]string
	for _, bbs := range re.FindAllSubmatch([]byte(s), n) {
		var ss []string
		for _, bs := range bbs {
			ss = append(ss, string(bs))
		}
		rs = append(rs, ss)
	}
	return rs
}

func (re *Regexp) FindAllStringSubmatchIndex(s string, n int) [][]int {
	return re.FindAllSubmatchIndex([]byte(s), n)
}

func (re *Regexp) FindAllSubmatch(b []byte, n int) [][][]byte {
	var rs [][][]byte
	for _, locs := range re.FindAllSubmatchIndex(b, n) {
		var bs [][]byte
		for i := 0; i < len(locs); i += 2 {
			if locs[i] == -1 || locs[i+1] == -1 {
				bs = append(bs, []byte{})
			} else {
				bs = append(bs, b[locs[i]:locs[i+1]])
			}
		}
		rs = append(rs, bs)
	}
	return rs
}

func (re *Regexp) FindAllSubmatchIndex(b []byte, n int) [][]int {
	var locs [][]int
	var options pcre.Option
	for start := 0; start <= len(b) && n != 0; n-- {
		ovector := make([]int, (1+re.pcre.Capturecount())*3)
		if e := re.pcre.Exec(nil, string(b), start, options, ovector); e == pcre.ErrNomatch {
			break
		} else if e < 0 {
			log.Panicf("while matching %q[%d:]: %d", b, start, e)
		} else {
			locs = append(locs, ovector[:(1+re.pcre.Capturecount())*2])
		}
		options |= pcre.NotemptyAtstart
		start = ovector[1]
	}

	return locs
}

func (re *Regexp) FindSubmatch(b []byte) [][]byte {
	var bs [][]byte
	locs := re.FindSubmatchIndex(b)
	for i := 0; i < len(locs); i += 2 {
		if locs[i] == -1 || locs[i+1] == -1 {
			bs = append(bs, []byte{})
		} else {
			bs = append(bs, b[locs[i]:locs[i+1]])
		}
	}
	return bs
}

func (re *Regexp) FindStringSubmatch(s string) []string {
	var subs []string
	for _, sub := range re.FindSubmatch([]byte(s)) {
		subs = append(subs, string(sub))
	}
	return subs
}

func (re *Regexp) FindSubmatchIndex(b []byte) []int {
	var t = string(b) == "aacc" || re.expr == "(a){0}"
	ovector := make([]int, (1+re.pcre.Capturecount())*3)
	if e := re.pcre.Exec(nil, string(b), 0, 0, ovector); e == pcre.ErrNomatch {
		return nil
	} else if e < 0 {
		log.Panicf("e: %s", e)
	} else if t {
		log.Printf("expr: %q, b: %q, e: %d, ovector: %#v, %#v", re.expr, b, e, ovector, ovector[:(1+re.pcre.Capturecount())*2])
	}
	return ovector[:(1+re.pcre.Capturecount())*2]
}

func (re *Regexp) FindStringSubmatchIndex(s string) []int {
	return re.FindSubmatchIndex([]byte(s))
}

func (re *Regexp) FindReaderSubmatchIndex(r io.RuneReader) []int {
	data, err := readAllRunes(r)
	if err != nil {
		log.Panicf("readAllRunes: %s", err)
	}

	return re.FindSubmatchIndex(data)
}

func (re *Regexp) LiteralPrefix() (prefix string, complete bool)                    { panic("TODO") }
func (re *Regexp) Longest()                                                         {} // TODO
func (re *Regexp) MatchReader(r io.RuneReader) bool                                 { panic("TODO") }

func (re *Regexp) Split(s string, n int) []string { return nil }

func (re *Regexp) String() string { return re.expr } // TODO

func (re *Regexp) NumSubexp() int {
	return re.pcre.Capturecount() // TODO - is this correct? or do they mean *all* groups?
}

func (re *Regexp) SubexpNames() []string { return re.pcre.Nametable() }

func readAllRunes(r io.RuneReader) ([]byte, error) {
	data := new(bytes.Buffer)
	for {
		if rune, _, err := r.ReadRune(); err != nil {
			break
		} else if _, err := data.WriteRune(rune); err != nil {
			return nil, err
		}
	}
	return data.Bytes(), nil
}
