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
import "github.com/maxymania/fastnntp-backend/posting"

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
func (a *Articledb) CheckPost() (possible bool) { return true }
func (a *Articledb) PerformPost(id []byte, r *fastnntp.DotReader) (rejected bool, failed bool) {
	head,body := posting.ConsumePostedArticle(r)
	if len(head)==0 || len(body)==0 { return false,true }
	return true,false
}

