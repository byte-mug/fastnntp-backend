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


package posting

import "github.com/maxymania/fastnntp"
import "bytes"
import "io"

var lineFeed = []byte("\n")

func cloneB(buf []byte) []byte {
	n := make([]byte,len(buf))
	copy(n,buf)
	return n
}

func toLower(b byte) byte {
	if (b >= 'A') && (b <= 'Z') { return (b-'A')+'a' }
	return b
}

func aToLower(buf []byte) {
	for i,b := range buf {
		buf[i] = toLower(b)
	}
}

func trimCRLF(b []byte) []byte {
	i := len(b)
	for i>0 {
		i--
		switch b[i] {
		case '\r','\n': continue
		default: return b[:i+1]
		}
	}
	return b[:0]
}
func trimDOT(b []byte) []byte {
	i := len(b)
	if i==0 { return b }
	i--
	if b[i]=='.' { return b[:i] }
	return b
}
func singleLineB(buf []byte) []byte {
	n := make([]byte,0,len(buf))
	for _,seg := range bytes.SplitAfter(buf,lineFeed) {
		n = append(n,trimCRLF(seg)...)
	}
	return n
}

func ConsumePostedArticle(r *fastnntp.DotReader) (head []byte, body []byte) {
	headw := new(bytes.Buffer)
	bodyw := new(bytes.Buffer)
	hbw := fastnntp.AcquireHeadBodyWriter()
	hbw.Reset(headw,bodyw)
	defer hbw.Release()
	dw := fastnntp.AcquireDotWriter()
	dw.Reset(hbw)
	defer func(){ dw.Close(); dw.Release() }()
	io.Copy(dw,r)
	
	head = trimCRLF(headw.Bytes())
	body = trimDOT(trimCRLF(bodyw.Bytes()))
	
	return
}

type HeadInfo struct{
	MessageId  []byte
	Newsgroups []byte
	Subject    []byte
	From       []byte
	Date       []byte
	References []byte
	
	// Raw header
	RAW        []byte
}

var standardHeaders = map[string]int {
	"message-id": 1,
	"newsgroups": 2,
	"subject"   : 3,
	"from"      : 4,
	"date"      : 5,
	"references": 6,
}
var headerCase = map[string][]byte {
	"message-id": []byte("Message-ID"),
	"newsgroups": []byte("Newsgroups"),
	"subject"   : []byte("Subject"),
	"from"      : []byte("From"),
	"date"      : []byte("Date"),
	"references": []byte("References"),
}

/*
func writeHeader(b *bytes.Buffer, cut int, line []byte ) {
	if cut>=76 { goto done }
	for len(line)>78 {
		i := bytes.LastIndexAny(line[:78]," \t")
		if i<cut { i=-1 }
		if i<0 {
			i := bytes.LastIndexAny(line[cut:78]," \t")
			if i>=0 { i+=cut }
		}
		if i<0 { i = 78 }
		b.Write(line[:i])
		b.WriteString("\r\n")
		line = line[i:]
	}
done:
	b.Write(line)
	b.WriteString("\r\n")
}
*/

func ParseAndProcessHeader(id []byte, s Stamper, head []byte) (hi *HeadInfo) {
	hi = new(HeadInfo)
	headw := new(bytes.Buffer)
	last := make([]byte,0,256)
	name := make([]byte,0,25)
	buffer := make([]byte,0,100)
	has_path := false
	has_id := false
	for _,el := range bytes.SplitAfter(head,lineFeed) {
		el = trimCRLF(el)
		if len(el)==0 { continue }
		switch el[0] {
		case ' ','\t':
			// For now, we preserve the Conatined Newline.
			last = append(last,"\r\n"...)
			last = append(last,el...); continue
		}
		if len(last)>0 {
			i := bytes.IndexByte(last,':')
			j := i+2
			unwrit := true
			if i>0 && i<25 {
				name = append(name[:0],last[:i]...)
				aToLower(name)
				val := last[j:]
				copy(last,headerCase[string(name)])
				switch standardHeaders[string(name)] {
				case 1: has_id = true
					if len(id)>0 && bytes.Equal(val,id) { return nil }
					hi.MessageId  = singleLineB(last[j:])
				case 2: hi.Newsgroups = singleLineB(last[j:])
				case 3: hi.Subject    = singleLineB(last[j:])
				case 4: hi.From       = singleLineB(last[j:])
				case 5: hi.Date       = singleLineB(last[j:])
				case 6: hi.References = singleLineB(last[j:])
				case 7:{
					has_path = true
					pb := s.PathSeg(buffer)
					if len(pb)>0 {
						unwrit = false
						headw.Write(last[:j])
						headw.Write(pb)
						headw.Write(last[:j])
						headw.WriteString("\r\n")
					}
				  }
				}
			}
			if unwrit {
				//writeHeader(headw,j,last)
				headw.Write(last)
				headw.WriteString("\r\n")
			}
		}
		last = append(last[:0],el...)
	}
	if !has_path {
		pb := s.PathSeg(buffer)
		if len(pb)>0 {
			headw.WriteString("Path: ")
			headw.Write(pb[:len(pb)-1])
			headw.WriteString("\r\n")
		}
	}
	if !has_id {
		idm := s.GetId(buffer)
		if len(idm)>0 {
			hi.MessageId  = cloneB(idm)
			headw.WriteString("Message-ID: ")
			headw.Write(idm)
			headw.WriteString("\r\n")
		}
	}
	hi.RAW = headw.Bytes()
	return
}

