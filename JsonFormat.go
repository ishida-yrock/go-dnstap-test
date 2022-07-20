/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

 package dnstap

 import (
	 "bytes"
	 "encoding/json"
	 "fmt"
	 "net"
	 "time"
 
	 "github.com/miekg/dns"
 )
 
 type jsonTime time.Time
 
 func (jt *jsonTime) MarshalJSON() ([]byte, error) {
	 stamp := time.Time(*jt).Format(time.RFC3339Nano)
	 return []byte(fmt.Sprintf("\"%s\"", stamp)), nil
 }
 
 type jsonDnstap struct {
	 Type     string      `json:"type"`
	 Identity string      `json:"identity,omitempty"`
	 Version  string      `json:"version,omitempty"`
	 Message  jsonMessage `json:"message"`
 }
 
 type jsonMessage struct {
	 Type            string    `json:"type"`
	 QueryTime       *jsonTime `json:"query_time,omitempty"`
	 ResponseTime    *jsonTime `json:"response_time,omitempty"`
	 SocketFamily    string    `json:"socket_family,omitempty"`
	 SocketProtocol  string    `json:"socket_protocol,omitempty"`
	 QueryAddress    *net.IP   `json:"query_address,omitempty"`
	 ResponseAddress *net.IP   `json:"response_address,omitempty"`
	 QueryPort       uint32    `json:"query_port,omitempty"`
	 ResponsePort    uint32    `json:"response_port,omitempty"`
	 QueryZone       string    `json:"query_zone,omitempty"`
	 QueryMessage    string    `json:"query_message,omitempty"`
	 ResponseMessage string    `json:"response_message,omitempty"`  //response_message単位で扱われている
 }
 
 func convertJSONMessage(m *Message) jsonMessage {
	 jMsg := jsonMessage{
		 Type:           fmt.Sprint(m.Type),
		 SocketFamily:   fmt.Sprint(m.SocketFamily),
		 SocketProtocol: fmt.Sprint(m.SocketProtocol),
	 }
 
	 if m.QueryTimeSec != nil && m.QueryTimeNsec != nil {
		 qt := jsonTime(time.Unix(int64(*m.QueryTimeSec), int64(*m.QueryTimeNsec)).UTC())
		 jMsg.QueryTime = &qt
	 }
 
	 if m.ResponseTimeSec != nil && m.ResponseTimeNsec != nil {
		 rt := jsonTime(time.Unix(int64(*m.ResponseTimeSec), int64(*m.ResponseTimeNsec)).UTC())
		 jMsg.ResponseTime = &rt
	 }
 
	 if m.QueryAddress != nil {
		 qa := net.IP(m.QueryAddress)
		 jMsg.QueryAddress = &qa
	 }
 
	 if m.ResponseAddress != nil {
		 ra := net.IP(m.ResponseAddress)
		 jMsg.ResponseAddress = &ra
	 }
 
	 if m.QueryPort != nil {
		 jMsg.QueryPort = *m.QueryPort
	 }
 
	 if m.ResponsePort != nil {
		 jMsg.ResponsePort = *m.ResponsePort
	 }
 
	 if m.QueryZone != nil {
		 name, _, err := dns.UnpackDomainName(m.QueryZone, 0)
		 if err != nil {
			 jMsg.QueryZone = fmt.Sprintf("parse failed: %v", err)
		 } else {
			 jMsg.QueryZone = string(name)
		 }
	 }
 
	 if m.QueryMessage != nil {
		 msg := new(dns.Msg)
		 err := msg.Unpack(m.QueryMessage)
		 if err != nil {
			 jMsg.QueryMessage = fmt.Sprintf("parse failed: %v", err)
		 } else {
			 jMsg.QueryMessage = msg.String()
		 }
	 }
 
	 if m.ResponseMessage != nil { //この中でjsonMarshal?
		 msg := new(dns.Msg)
		 //1 variable but json.Marshal return 2 valuesのエラー。多分元々response_messageと値が1:1だったのを、中身をいくつもjson化してるから起こった？配列とかに順にいれさせて吐かせられないか？
		 err := msg.Unpack(m.ResponseMessage) // エラー処理
		 if err != nil {
			 jMsg.ResponseMessage = fmt.Sprintf("parse failed: %v", err)
		 } else {
			 jMsg.ResponseMessage = msg.String() //こいつがresponse_messageの正体で、その中身を:と,ごとにさらにjson化したい  一個ずつopcodeとかも構造化に入れて定義する？
		 } //構造化定義でjson:response_message,omitemptyと定義されているので、標準出力で「"response_message":"~~~~~"」とreponseに対して大量の平文を1:1に表示される
	 }
	 return jMsg  //jsonMessage構造のjson文字列を返す
 }
 
 // JSONFormat renders a Dnstap message in JSON format. Any encapsulated
 // DNS messages are rendered as strings in a format similar to 'dig' output.
 func JSONFormat(dt *Dnstap) (out []byte, ok bool) {
	 var s bytes.Buffer
 
	 j, err := json.Marshal(jsonDnstap{ //json.Marshalで構造体をjsonに変換
		 Type:     fmt.Sprint(dt.Type),  //sprint:文字列と変数を組みあわせて新しい文字列を作る
		 Identity: string(dt.Identity),  
		 Version:  string(dt.Version),
		 Message:  convertJSONMessage(dt.Message),
	 })
	 if err != nil {
		 return nil, false
	 }
 
	 s.WriteString(string(j) + "\n")
 
	 return s.Bytes(), true
 }
 