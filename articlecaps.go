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

import "github.com/byte-mug/fastnntp"
import "github.com/vmihailenco/msgpack"

type articleMetadata struct{
	Refc int64
	Nums map[string]int64
}

type articleOver struct{
	SF [5][]byte // Subject, From, Date, MsgId, Refs
	LN [2]int64  // Bytes, Lines
}

//func (a *Articledb) 
//func (a *articleTransaction) 

func (a *Articledb) StatArticle(ar *fastnntp.Article) bool {
	t,e := a.begin(false)
	if e!=nil { return false }
	defer t.commit()
	return t.StatArticle(ar)
}

func (a *articleTransaction) StatArticle(ar *fastnntp.Article) bool {
	if ar.HasNum && !ar.HasId {
		bkt := a.tx.Bucket(tGRPARTS).Bucket(ar.Group)
		if bkt==nil { return false }
		v := bkt.Get(encode64(ar.Number))
		if len(v)==0 { return false }
		ar.MessageId = append(ar.MessageId[:0],v...)
		ar.HasId = true
		return true
	}
	if ar.HasId {
		if len(a.tx.Bucket(tARTMETA).Get(ar.MessageId))==0 { return false }
		return true
	}
	return false
}

// TODO: Optimize this allocation away!
func (a *Articledb) getArticleClosure(ar *fastnntp.Article, head, body bool) func(w *fastnntp.DotWriter) {
	return func(w *fastnntp.DotWriter) { a.getArticle(ar, head, body, w) }
}


func (a *Articledb) getArticle(ar *fastnntp.Article,head, body bool, w *fastnntp.DotWriter) {
	t,e := a.begin(false)
	if e!=nil { return }
	defer t.commit()
	t.getArticle(ar, head, body, w)
}
func (a *articleTransaction) getArticle(ar *fastnntp.Article,head, body bool, w *fastnntp.DotWriter) {
	bid := ar.MessageId
	if head {
		h := a.tx.Bucket(tARTHEAD).Get(bid)
		w.Write(h)
	}
	if head && body { w.Write([]byte("\r\n")) }
	if body {
		h := a.tx.Bucket(tARTBODY).Get(bid)
		w.Write(h)
	}
	w.Write([]byte(".\r\n"))
}
func (a *Articledb) GetArticle(ar *fastnntp.Article, head, body bool) func(w *fastnntp.DotWriter) {
	if !a.StatArticle(ar) { return nil }
	return a.getArticleClosure(ar, head, body)
}


func (a *Articledb) writeOverviewId(ar *fastnntp.ArticleRange, w fastnntp.IOverview) {
	t,e := a.begin(false)
	if e!=nil { return }
	defer t.commit()
	t.writeOverviewId(ar, w)
}
func (a *articleTransaction) writeOverviewId(ar *fastnntp.ArticleRange, w fastnntp.IOverview) {
	var ao articleOver
	v := a.tx.Bucket(tARTOVER).Get(ar.MessageId)
	if msgpack.Unmarshal(v,&ao)!=nil { return }
	w.WriteEntry(0, ao.SF[0], ao.SF[1], ao.SF[2], ao.SF[3], ao.SF[4], ao.LN[0], ao.LN[1])
}
func (a *Articledb) writeOverviewRange(ar *fastnntp.ArticleRange, w fastnntp.IOverview) {
	t,e := a.begin(false)
	if e!=nil { return }
	defer t.commit()
	t.writeOverviewRange(ar, w)
}
func (a *articleTransaction) writeOverviewRange(ar *fastnntp.ArticleRange, w fastnntp.IOverview) {
	var ao articleOver
	bkt := a.tx.Bucket(tGRPARTS).Bucket(ar.Group)
	if bkt==nil { return }
	c := bkt.Cursor()
	k,m := c.Seek(encode64(ar.Number))
	for ; len(k)!=0 ; k,m = c.Next() {
		num := decode64(k)
		if num>ar.LastNumber { break }
		v := a.tx.Bucket(tARTOVER).Get(m)
		if msgpack.Unmarshal(v,&ao)!=nil { continue }
		w.WriteEntry(num, ao.SF[0], ao.SF[1], ao.SF[2], ao.SF[3], ao.SF[4], ao.LN[0], ao.LN[1])
	}
}

func (a *Articledb) WriteOverview(ar *fastnntp.ArticleRange) func(w fastnntp.IOverview) {
	if ar.HasId {
		if !a.StatArticle(&(ar.Article)) { return nil }
		return func(w fastnntp.IOverview) { a.writeOverviewId(ar,w) }
	}
	if ar.HasNum {
		return func(w fastnntp.IOverview) { a.writeOverviewRange(ar,w) }
	}
	return nil
}

