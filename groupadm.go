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
import "github.com/vmihailenco/msgpack"

//func (a *Articledb) 
//func (a *articleTransaction) 

func (a *Articledb) AdmAddGroup(group, descr []byte) error {
	t,e := a.begin(true)
	if e!=nil { return e }
	defer t.commit()
	return t.AdmAddGroup(group,descr)
}
func (a *articleTransaction) AdmAddGroup(group, descr []byte) error {
	gnums,e := msgpack.Marshal(&groupInfo{0,0,0,'n'})
	if e!=nil { return e }
	v := a.tx.Bucket(tGRPINFO).Get(group)
	if len(v)!=0 { return fmt.Errorf("Group exists %v",string(group)) }
	v = a.tx.Bucket(tGRPNUMS).Get(group)
	if len(v)!=0 { return fmt.Errorf("Group exists %v",string(group)) }
	a.tx.Bucket(tGRPINFO).Put(group,descr)
	a.tx.Bucket(tGRPNUMS).Put(group,gnums)
	a.tx.Bucket(tGRPARTS).CreateBucketIfNotExists(group)
	return nil
}

