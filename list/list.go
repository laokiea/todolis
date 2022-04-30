package list

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"syscall"

	"github.com/fatih/color"
)

type List interface {
	List() []*ListItem
	Add(string)
	Done(int)
	Del(int)
	Len()
	Search(string) []ListItem
}

type ListMap struct {
	m       []*ListItem
	len     int
	p       *regexp.Regexp
	lastKey string
}

type ListItem struct {
	v    string
	done bool
}

func (i ListItem) String() string {
	return i.v
}

func (i ListItem) Status() string {
	if i.done {
		return fmt.Sprintf("[%s]", color.GreenString("done"))
	}
	return fmt.Sprintf("[%s]", color.RedString("undone"))
}

func (i ListItem) Display() string {
	return fmt.Sprintf("%s %s", i.Status(), color.New(color.FgHiBlack).Add(color.Bold).Add(color.BgWhite).Sprint(i.String()))
}

func (l *ListMap) Len() int {
	return len(l.m)
}

func (l *ListMap) List() []*ListItem {
	return l.m
}

func (l *ListMap) ListSliceAll() []string {
	var s []string
	for _, i := range l.m {
		s = append(s, i.Display())
	}
	return s
}

func (l *ListMap) ListSliceUndone() []string {
	var s []string
	for _, i := range l.m {
		if !i.done {
			s = append(s, i.Display())
		}
	}
	return s
}

func (l *ListMap) Add(v string) {
	l.AddWithDone(v, false)
}

func (l *ListMap) AddWithDone(v string, done bool) {
	l.len++
	l.m = append(l.m, &ListItem{v: v, done: done})
}

func (l *ListMap) Done(i int) {
	l.m[i].done = true
}

func (l *ListMap) Del(i int) {
	l.len--
	l.m = append(l.m[:i], l.m[i+1:]...)
}

func (l *ListMap) Search(k string) []*ListItem {
	var lists []*ListItem
	if k != l.lastKey {
		l.lastKey = k
		l.p = regexp.MustCompile(`^.*` + k + `.*$`)
	}
	for _, i := range l.m {
		if l.p.MatchString(i.v) {
			lists = append(lists, i)
		}
	}
	return lists
}

func (l *ListMap) Flush() error {
	syscall.Unlink("./todolist")
	f, err := os.OpenFile("./todolist", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	var done string
	for _, i := range l.m {
		if i.done {
			done = "1"
		} else {
			done = "0"
		}
		_, err = f.Write([]byte(fmt.Sprintf("%s|%s\n", i.v, done)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *ListMap) Load() error {
	f, err := os.OpenFile("./todolist", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		pos := strings.LastIndex(line, "|")
		done := false
		if line[pos+1:] == "1" {
			done = true
		}
		l.AddWithDone(line[:pos], done)
	}
	return nil
}

var (
	GlobalLists ListMap
)
