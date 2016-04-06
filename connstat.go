package main
import (
 "fmt"
 "os"
 "io"
 "io/ioutil"
 //"path/filepath"
 "strings"
 "strconv"
 "bufio"
 "net"
 "sort"
)
//-------------------------------------------------------------------------------------------------
const root string="/proc/";
var local_ips=[]string{};
//-------------------------------------------------------------------------------------------------

type TCPInfor struct{
 ip string;
 port uint16;
 in bool;
};

//-------------------------------------------------------------------------------------------------
type TIS []TCPInfor;
func (l TIS) Len() int {
 return len(l);
}
func (l TIS) Swap(i,j int){
 l[i],l[j]=l[j],l[i];
}
func (l TIS) Less(i,j int)bool{
 if l[i].ip==l[j].ip {
    return l[i].port<l[j].port;
 }
 return l[i].ip<l[j].ip;
}
//-------------------------------------------------------------------------------------------------

func get_local_ip()([]string,error){
 ifaces,err:=net.Interfaces()
 if err!=nil {
    return nil,err;
 }
 var ips []string;
 for _,i:=range ifaces {
     addrs,err:=i.Addrs();
     if err!=nil {
        return nil,err;
     }
     for _,addr:=range addrs {
         var ip net.IP;
         switch v:= addr.(type) {
         case *net.IPNet:
              ip=v.IP;
         case *net.IPAddr:
              ip=v.IP;
         }//end switch
	 if ip.String()!="127.0.0.1"{
	    ips=append(ips,ip.String());
	 }
     }
 }//end for i
 return ips,nil;
}
//-------------------------------------------------------------------------------------------------

func is_local_ip(ip string)(bool){
 for _,lip:=range local_ips {
     if lip==ip {
        return true;
     }
 }
 return false;
}
//-------------------------------------------------------------------------------------------------

func get_all_in_out_connections()([]TCPInfor,[]TCPInfor,error){
 in:=[]TCPInfor{};
 out:=[]TCPInfor{};
 dirs,err:=ioutil.ReadDir(root);
 if err!=nil{
    return nil,nil,err;
 }
 for _,fi:=range dirs {
  var pn=0;
  pn,err:=strconv.Atoi(fi.Name());
  if err!=nil||pn<2 {
     continue;
  }
 var filename="";
 /* 
 var filename=root+fi.Name()+string(os.PathSeparator)+"cmdline";
 fmt.Println(filename);
 fh,err:=os.Open(filename);
 if err!=nil {
    continue;
 }
 fi,err:=fh.Stat();
 fh.Close();
 if fi.Size()<=0 {
     continue;
 }*/

  filename=root+fi.Name()+string(os.PathSeparator)+"net"+string(os.PathSeparator)+"tcp";
  fmt.Printf("parsing:");
  fmt.Println(filename);
  in,out,err=parse_in_out_connection_of_process(filename,in,out);
  //break;
 }
 return in,out,nil;
}
//-------------------------------------------------------------------------------------------------

func get_connection_infor(line string)(TCPInfor){
 var ti TCPInfor;
 var items=strings.Fields(line);
 if len(items)<5 {
    fmt.Printf("items error %d\n",len(items));
    return ti;
 }
 var t1=strings.Split(items[1],":");
 var t2=strings.Split(items[2],":");
 if len(t1)!=2||len(t2)!=2 {
    //fmt.Printf("split error [%s,%s]\n",items[1],items[2]);
    return ti;
 }
 var from_addr=[4]byte{0,0,0,0};
 var to_addr=[4]byte{0,0,0,0};
 var status=items[3];
 if status!="01"{
    return ti;
 }
 fmt.Sscanf(t1[0],"%02X%02X%02X%02X",
            &from_addr[3],&from_addr[2],&from_addr[1],&from_addr[0]);
 if from_addr[0]==0&&from_addr[1]==0&&from_addr[2]==0&&from_addr[3]==0 {
    return ti;
 }
 if(from_addr[0]==127&&from_addr[1]==0&&from_addr[2]==0&&from_addr[3]==1) {
    return ti;
 }
 var from_addr_string=fmt.Sprintf("%d.%d.%d.%d",from_addr[0],from_addr[1],from_addr[2],from_addr[3]);

 fmt.Sscanf(t2[0],"%02X%02X%02X%02X",
            &to_addr[3],&to_addr[2],&to_addr[1],&to_addr[0]);
 if to_addr[0]==0&&to_addr[1]==0&&to_addr[2]==0&&to_addr[3]==0 {
    return ti;
 }
 if(to_addr[0]==127&&to_addr[1]==0&&to_addr[2]==0&&to_addr[3]==1) {
    return ti;
 }
 var to_addr_string=fmt.Sprintf("%d.%d.%d.%d",to_addr[0],to_addr[1],to_addr[2],to_addr[3]);

 var from_port=0;
 fmt.Sscanf(t1[1],"%04X",&from_port);
 var to_port=0;
 fmt.Sscanf(t2[1],"%04X",&to_port);
 //fmt.Println(from_addr,from_port,to_addr,to_port);

 if is_local_ip(from_addr_string) {
     ti.in=false;
     ti.ip=to_addr_string;
     ti.port=(uint16)(to_port);
 } else {
     ti.in=true;
     ti.ip=from_addr_string;
     ti.port=(uint16)(to_port);
 }
 return ti;
}
//-------------------------------------------------------------------------------------------------

func parse_in_out_connection_of_process(filename string,in []TCPInfor,
                                        out []TCPInfor)([]TCPInfor,[]TCPInfor,error){
 fh,err:=os.Open(filename);
 if err!=nil {
    return in,out,err;
 }
 defer fh.Close();
 reader:=bufio.NewReader(fh);
 var k=0;
 for {
  k++;
  if k>=5000 {
     break;
  }
  line,err:=reader.ReadString('\n');
  if err!=nil||err==io.EOF {
   break;
  }
  line=strings.TrimSpace(line);
  if line!=""{
     ti:=get_connection_infor(line);
     if ti.port==0 {
        continue;
     }
     if ti.in==true {
        in=append(in,ti);
     } else {
        out=append(out,ti);
     }
  }
 }//end for
 return in,out,err;
}
//-------------------------------------------------------------------------------------------------

func main() {
 var in_array TIS;
 var out_array TIS;
 var err error;

 local_ips,err=get_local_ip();
 in_array,out_array,err=get_all_in_out_connections();
 if err!=nil{
  fmt.Println(err);
  return;
 }
 //fmt.Println(in_array);
 //fmt.Println(out_array);
 sort.Sort(in_array);
 sort.Sort(out_array);

 var in_array2 TIS;
 var out_array2 TIS;
 for i,ti:=range in_array {
     if i>0&&ti==in_array[i-1] {
        continue;
     }
     in_array2=append(in_array2,ti);
 }
 for i,ti:=range out_array {
     if i>0&&ti==out_array[i-1] {
        continue;
     }
     out_array2=append(out_array2,ti);
 }
 
 var in_map=make(map[string]int);
 var out_map=make(map[string]int);
 for _,ti:=range in_array2{
     v,exists:=in_map[ti.ip];
     if exists==true {
        in_map[ti.ip]=v+1;
     } else {
        in_map[ti.ip]=1;
     }
 }
 for _,ti:=range out_array2{
     v,exists:=out_map[ti.ip];
     if exists==true {
        out_map[ti.ip]=v+1;
     } else {
        out_map[ti.ip]=1;
     }
 }
 fmt.Println("\n-------------------------------------\n");
 fmt.Println("in connections:");
 fmt.Println(in_map);
 fmt.Println("\n-------------------------------------\n");
 fmt.Println("out connections:");
 fmt.Println(out_map);
}
//-------------------------------------------------------------------------------------------------
