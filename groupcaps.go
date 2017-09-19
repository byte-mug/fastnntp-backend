/*
MIT License

Copyright (c) 2017 Simon Schmidt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/


package fnntpbackend

import "fmt"
import "github.com/maxymania/fastnntp"
import "github.com/vmihailenco/msgpack"

type groupInfo [4]int64 /* [ Count, Low, High, Status ]*/

//func (a *Articledb) 
//func (a *articleTransaction) 

func (a *Articledb) GetGroup(g *fastnntp.Group) bool {
	t,e := a.begin(false)
	if e!=nil { return false }
	defer t.commit()
	return t.GetGroup(g)
}

func (a *articleTransaction) getGroup(group []byte) (gi groupInfo,ok bool) {
	b := a.tx.Bucket(tGRPNUMS).Get(group)
	if len(b)==0 { return }
	e := msgpack.Unmarshal(b,&gi)
	ok = e==nil
	return
}
func (a *articleTransaction) GetGroup(g *fastnntp.Group) bool {
	gi,ok := a.getGroup(g.Group)
	if !ok { return false }
	g.Number = gi[0]
	g.Low    = gi[1]
	g.High   = gi[2]
	return true
}

func (a *Articledb) ListGroup(g *fastnntp.Group, w *fastnntp.DotWriter, first, last int64) {
	t,e := a.begin(false)
	if e!=nil { return }
	defer t.commit()
	t.ListGroup(g,w,first,last)
}
func (a *articleTransaction) ListGroup(g *fastnntp.Group, w *fastnntp.DotWriter, first, last int64) {
	/*
	gi,ok := a.getGroup(g.Group)
	if !ok { return }
	if first < gi[1] { first = gi[1] }
	if last  > gi[2] { last  = gi[2] }
	*/
	bkt := a.tx.Bucket(tGRPARTS).Bucket(g.Group)
	if bkt==nil { return }
	c := bkt.Cursor()
	k,_ := c.Seek(encode64(first))
	for len(k)>0 {
		num := decode64(k)
		if num>last { break }
		fmt.Fprintf(w,"%v\r\n",num)
		k,_ = c.Next()
	}
}

func (a *Articledb) CursorMoveGroup(g *fastnntp.Group, i int64, backward bool, id_buf []byte) (ni int64, id []byte, ok bool) {
	t,e := a.begin(false)
	if e!=nil { ok = false; return }
	defer t.commit()
	return t.CursorMoveGroup(g,i,backward,id_buf)
}
func (a *articleTransaction) CursorMoveGroup(g *fastnntp.Group, i int64, backward bool, id_buf []byte) (ni int64, id []byte, ok bool) {
	bkt := a.tx.Bucket(tGRPARTS).Bucket(g.Group)
	if bkt==nil { return }
	c := bkt.Cursor()
	k,v := c.Seek(encode64(i))
	if len(k)==0 { return }
	if backward { k,v = c.Prev() } else { k,v = c.Next() }
	if len(k)==0 { return }
	
	ni = decode64(k)
	id = append(id_buf,v...)
	ok = true
	return
}

