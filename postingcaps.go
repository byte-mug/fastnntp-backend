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

import "github.com/maxymania/fastnntp"
import "github.com/maxymania/fastnntp/posting"
import "github.com/vmihailenco/msgpack"

//func (a *Articledb) 
//func (a *articleTransaction) 

func (a *articleTransaction) CheckPostId(id []byte) (wanted bool, possible bool) {
	v := a.tx.Bucket(tARTMETA).Get(id)
	if len(v)>0 { return false,true }
	return true,true
}


func (a *Articledb) CheckPostId(id []byte) (wanted bool, possible bool) {
	t,e := a.begin(false)
	if e!=nil { return true,false }
	defer t.commit()
	return t.CheckPostId(id)
}


// TODO: Replace.
var hostname = posting.HostName("localhost")

func (a *Articledb) CheckPost() (possible bool) { return true }
func (a *Articledb) PerformPost(id []byte, r *fastnntp.DotReader) (rejected bool, failed bool) {
	head,body := posting.ConsumePostedArticle(r)
	if len(head)==0 || len(body)==0 { return false,true }
	
	// TODO: Replace ''hostname''
	headp := posting.ParseAndProcessHeader(id,hostname,head)
	if headp==nil { return true,false }
	
	if len(headp.MessageId)==0 { return false,true } // no message-ID? Failed.
	
	ngrps := posting.SplitNewsgroups(headp.Newsgroups)
	if len(ngrps)==0 { return true,false }
	
	t,e := a.begin(true)
	if e!=nil { return false,true }
	defer t.commit()
	return t.performPost(headp,body,ngrps)
}
func (a *articleTransaction) filterNewsgroups(ngs [][]byte) [][]byte {
	i := 0
	nums := a.tx.Bucket(tGRPNUMS)
	gi := new(groupInfo)
	
	for _,group := range ngs {
		v := nums.Get(group)
		if len(v)==0 { continue }
		if msgpack.Unmarshal(v,gi)!=nil { continue }
		if gi[3]!='y' { continue }
		ngs[i] = group
		i++
	}
	
	return ngs[:i]
}
func (a *articleTransaction) insertIntoGroups(id []byte, am *articleMetadata, ngs [][]byte) {
	nums := a.tx.Bucket(tGRPNUMS)
	arts := a.tx.Bucket(tGRPARTS)
	gi := new(groupInfo)
	
	for _,group := range ngs {
		if _,ok := am.Nums[string(group)] ; ok { continue } // avoid duplicate insertion
		
		
		v := nums.Get(group)
		if len(v)==0 { continue }
		if msgpack.Unmarshal(v,gi)!=nil { continue }
		
		gi[0]++ // Number
		if gi[1]==0 { gi[1] = 1 } // Low
		gi[2]++ // High
		num := gi[2]
		
		gbk := arts.Bucket(group)
		if gbk==nil { continue } // Can't update group
		if gbk.Put(encode64(num),id)!=nil { continue }
		
		
		v,_ = msgpack.Marshal(gi)
		nums.Put(group,v)
		am.Nums[string(group)] = num // Set number.
		am.Refc++
	}
}
func (a *articleTransaction) performPost(headp *posting.HeadInfo,body []byte, ngs [][]byte) (rejected bool, failed bool) {
	ngs = a.filterNewsgroups(ngs)
	
	if len(ngs)==0 { a.rollback(); return true,false }
	
	am := &articleMetadata{ Nums: make(map[string]int64) }
	ao := &articleOver{}
	
	// Subject, From, Date, MsgId, Refs
	ao.SF[0] = headp.Subject
	ao.SF[1] = headp.From
	ao.SF[2] = headp.Date
	ao.SF[3] = headp.MessageId
	ao.SF[4] = headp.References
	
	// Bytes, Lines
	ao.LN[0] = int64(len(headp.RAW)+2+len(body))
	ao.LN[1] = posting.CountLines(body)
	
	aob,_ := msgpack.Marshal(ao)
	
	a.tx.Bucket(tARTOVER).Put(headp.MessageId,aob)
	a.tx.Bucket(tARTHEAD).Put(headp.MessageId,headp.RAW)
	a.tx.Bucket(tARTBODY).Put(headp.MessageId,body)
	
	a.insertIntoGroups(headp.MessageId,am,ngs)
	if am.Refc==0 { a.rollback(); return true,false }
	
	amb,_ := msgpack.Marshal(am)
	a.tx.Bucket(tARTMETA).Put(headp.MessageId,amb)
	return
}

