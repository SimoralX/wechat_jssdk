package main

import (
	"fmt"		
	"net/http"
	"net/url"
	"sync"			
	"github.com/gorilla/mux"
	"reflect"  // IsEmpty'
	"time"  // createTimestamp
	"log"  // log.Println
	"math/rand"  // 获取随机数  
	"strconv"  // createNonceStr
	"crypto/sha1"  // sha1
	"encoding/hex"  // sha1
	"encoding/json"
	"os"
)


/*
	缓存access_token的Map，map包含内容：
	@params "token" access_token
	@params "time" 上次请求时的时间
	@params "expires_in" 有效时间时间间隔
*/
var map_token = make(map[string]string)
/*
	缓存ticket的Map，map包含内容：
	@params "ticket" ticket
	@params "time" 上次请求时的时间
	@params "expires_in" 有效时间时间间隔
*/
var map_ticket = make(map[string]string)

var m_token *Token
var m_ticket *Ticket
var mu_token sync.Mutex
var mu_ticket sync.Mutex



// 获取token的url
const TokenUrl = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential"
// appid
const Appid = "**********************************"
// secret
const AppSecret = "********************************"
// expires_in
const ExpiresInTemp = "7200"
// 离2小时还有20分钟的时候刷新
const EndTime = 1200
// 获取ticket的url
const TicketUrl = "https://api.weixin.qq.com/cgi-bin/ticket/getticket?type=jsapi"
// 随机字符串长度
const NonceStrLength = 15

// token单例模式
type Token struct {	
	Time string
	AccessToken string
	ExpiresIn string	
}

// ticket单例模式
type Ticket struct {	
	Time string
	Ticket string
	ExpiresIn string	
}

// token响应类型
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
}

// ticket响应类型
type TicketResponse struct {
	Ticket string `json:"ticket"`
	ExpiresIn int `json:"expires_in"`		
}

// APP返回类型
type AppResponse struct {
	Rel bool `json:"rel"`
	// Msg成员的Tag还带了一个额外的omitempty选项，表示当Go语言结构体成员为空或零值时不生成该JSON对象（这里false为零值）
	Msg string `json:"msg,omitempty"`  
	Result AppResult `json:"result,omitempty"`
}
// APP返回类型里面结果
type AppResult struct {
	EffectTime string `json:"effecttime,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	Ticket string `json:"ticket,omitempty"`
	TimeStamp string `json:"timestamp,omitempty"`
	NonceStr string `json:"nonceStr,omitempty"`
	Signature string `json:"signature,omitempty"`
	Appid string `json:"appid,omitempty"`	
}




// 主函数
func main() {	
	log.Println("服务启动")
	fmt.Println()	

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	wd += "/www/"	
	log.Println("静态资源地址", wd)

	// [CSDN](https://blog.csdn.net/wangshubo1989/article/details/71128972?utm_source=copy) 
	router := mux.NewRouter()    
	router.HandleFunc("/jssdk/gettoken", GetToken).Methods("GET")	  
	router.HandleFunc("/jssdk/getticket", GetTicket).Methods("GET")
	router.HandleFunc("/jssdk/getconfig", GetConfig).Methods("GET")
	// This will serve files under http://localhost:8000/static/<filename>
	// the truly file in path: /www/
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(wd))))
    
	log.Fatal(http.ListenAndServe(":80", router))	
}



/* #region token */
// GetAccessToken echoes the Path component of the request URL r.
func GetToken(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	log.Println("***  GetToken()，请求参数", r)

	if err := r.ParseForm(); err != nil {
        log.Println(err)
	}
		
	var response AppResponse
	  
	instance := GetTokenInstance()			

	if IsEmpty(instance.AccessToken) {		
		response.Rel = false
		response.Msg = "get access_token fail"			
	} else {		
		response.Rel = true
		response.Result.AccessToken = instance.AccessToken	
	}
	log.Println("*** 返回结果：", response, instance.AccessToken)		
	fmt.Println()
	json.NewEncoder(w).Encode(response)   
}

// 单例模式获取access_token
func GetTokenInstance() *Token {	    				
	timeStartStr, ok := map_token["time"]   
	if ok {
		log.Println("缓存time：", timeStartStr)						
	} else {
		log.Println("缓存time不存在请直接重新获取access_token")
	}
	timeStartInt, err := strconv.ParseInt(timeStartStr, 10, 64)  // string ==> int64
	if err != nil {
		log.Println("时间转换失败")
	}
	token, ok := map_token["token"]
	if ok {
		log.Println("缓存access_token：", token)
	} else {
		log.Println("缓存access_token不存在请直接重新获取access_token")
	}
	expiresInStr, ok := map_token["expires_in"]
	if ok {
		log.Println("缓存expires_in：", expiresInStr)
	} else {
		log.Println("缓存expires_in不存在，直接使用7200")
		expiresInStr = ExpiresInTemp
	}
	expiresInInt, err := strconv.Atoi(expiresInStr)  // string ==> int
	if err != nil {
		log.Println("时间转换失败")
	}

	// 获取时间
	timeNow := CreateTimestamp()
	
	log.Println("token", token)
	log.Println("timeStartInt", timeStartInt)
	log.Println("timeNow", timeNow)
	log.Println("expiresInInt", expiresInInt)
	log.Println("EndTime", EndTime)

	if IsEmpty(token) || token=="" || IsEmpty(timeStartInt) || timeStartInt==int64(0) || timeNow-timeStartInt >= int64(expiresInInt-EndTime) {
		mu_token.Lock()  // add lock
		log.Println("accessToken超时，或者不存在，重新获取");			
		result, err := GetNewToken()
		if err != nil {
			log.Fatal(err)
		} else {
			map_token["time"] = strconv.FormatInt(timeNow, 10)
			map_token["token"] = result.AccessToken
			map_token["expires_in"] = strconv.Itoa(result.ExpiresIn)
		}
		defer mu_token.Unlock()  // unlock
	}

	m_token = &Token {
		Time: map_token["time"],
		AccessToken: map_token["token"],
		ExpiresIn: map_token["expires_in"],
	}						            
    return m_token
}

// 获取新token
func GetNewToken() (*TokenResponse, error){
	/*
	* 	获取access_token
	* 	[参考网址](https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140183)
	*	[go参考网址](http://docscn.studygolang.com/pkg/net/url/#Values)
	*	https请求方式: GET
	* 	https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=APPID&secret=APPSECRET	
	*/	
	appid := url.QueryEscape(Appid)
	secret := url.QueryEscape(AppSecret)	

	resp, err := http.Get(TokenUrl + "&appid=" + appid + "&secret=" + secret)
	if err != nil {
		return nil, err
	}	

	// We must close resp.Body on all execution paths.
	// (Chapter 5 presents 'defer', which makes this simpler.)
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("search query failed: %s", resp.Status)
	}

	var result TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		resp.Body.Close()
		return nil, err
	}
	
	return &result, nil
}
/* #endregion */





/* #region ticket */
// GetAccessToken echoes the Path component of the request URL r.
func GetTicket(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	log.Println("*** GetTicket()，请求参数", r)

	if err := r.ParseForm(); err != nil {
        log.Println(err)
	}	
	var response AppResponse
	

	// 获取access_token
	instance := GetTicketInstance()			
	if IsEmpty(instance.Ticket) {		
		response.Rel = false
		response.Msg = "get ticket fail"			
	} else {		
		response.Rel = true
		response.Result.Ticket = instance.Ticket	
		response.Result.EffectTime = instance.Time
	}
	log.Println("*** 返回结果：", response, instance.Ticket)					
	fmt.Println()
	json.NewEncoder(w).Encode(response)   
}

// 单例模式获取ticket
func GetTicketInstance() *Ticket {	    
	log.Println("GetTicketInstance()，进入单例")
	timeStartStr, ok := map_ticket["time"]   
	if ok {
		log.Println("缓存time,timeStartStr:", timeStartStr)						
	} else {
		log.Println("缓存time不存在请直接重新获取ticket")
	}
	timeStartInt, err := strconv.ParseInt(timeStartStr, 10, 64)  // string ==> 64
	if err != nil {
		log.Println("时间转换失败,timeStartInt", timeStartInt)
	}
	ticket, ok := map_ticket["ticket"]
	if ok {
		log.Println("缓存ticket：", ticket)
	} else {
		log.Println("缓存ticket不存在请直接重新获取ticket")
	}
	expiresInStr, ok := map_ticket["expires_in"]
	if ok {
		log.Println("缓存expires_in：", expiresInStr)
	} else {
		log.Println("缓存expires_in不存在，直接使用7200")
		expiresInStr = ExpiresInTemp
	}
	expiresInInt, err := strconv.Atoi(expiresInStr)  // string ==> int
	if  err != nil {
		log.Println("时间转换失败,expiresInInt", expiresInInt)
	}

	// 获取时间
	timeNow := CreateTimestamp()
	
	log.Println("ticket", ticket)
	log.Println("timeStartInt", timeStartInt)
	log.Println("timeNow", timeNow)
	log.Println("expiresInInt", expiresInInt)
	log.Println("EndTime", EndTime)

	if IsEmpty(ticket) || ticket=="" || IsEmpty(timeStartInt) || timeStartInt==int64(0)  || timeNow-timeStartInt >= int64(expiresInInt-EndTime) {
		mu_ticket.Lock()
		log.Println("重新获取ticket");
		instance := GetTokenInstance()  // 实现token单例获取
		result, err := GetNewTicket(instance.AccessToken)
		if err != nil {
			log.Fatal("获取新ticket失败", err)
		} else {				
			map_ticket["time"] = strconv.FormatInt(timeNow, 10)
			map_ticket["ticket"] = result.Ticket
			map_ticket["expires_in"] = strconv.Itoa(result.ExpiresIn)
		}
		defer mu_ticket.Unlock()
	}
	m_ticket = &Ticket {
		Time: map_ticket["time"],
		Ticket: map_ticket["ticket"],
		ExpiresIn: map_ticket["expires_in"],
	}						        
    return m_ticket
}

// 获取新ticket
func GetNewTicket(token string) (*TicketResponse, error){
	/*
	* 	获取ticket
	* 	[参考网址](https://mp.weixin.qq.com/wiki?action=doc&id=mp1421141115&t=0.2169902626207505&token=&lang=zh_CN#2)	
	*/	
	access_token := url.QueryEscape(token)
	resp, err := http.Get(TicketUrl + "&access_token=" + access_token)
	if err != nil {
		return nil, err
	}	
	// We must close resp.Body on all execution paths.
	// (Chapter 5 presents 'defer', which makes this simpler.)
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("search query failed: %s", resp.Status)
	}
	var result TicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		resp.Body.Close()
		return nil, err
	}
	return &result, nil
}
/* #endregion */




/* #region 获取签名 */
func GetConfig(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	log.Println("***  GetConfig()，请求参数", r)
	if err := r.ParseForm(); err != nil {
        log.Println(err)
	}
	url := ""  // 请求的url    
	if r.Form["url"] != nil {
		url = r.Form["url"][0]
		log.Println("请求url：", url)		
	}
	var response AppResponse
	if !IsEmpty(url) {   
		// 获取access_token
		instance := GetTicketInstance()			
		if IsEmpty(instance.Ticket) {
			log.Println("ticket获取失败")		
			response.Rel = false
			response.Msg = "get ticket fail"			
		} else {
			log.Println("ticket获取成功")		
			// jsapi_ticket=sM4AOVdWfPE4DxkXGEs8VMCPGGVi4C3VM0P37wVUCFvkVAy_90u5h9nbSlYy3-Sl-HhTdfl2fzFy1AOcHKP7qg&noncestr=Wm3WZYTPz0wzccnW&timestamp=1414587457&url=http://mp.weixin.qq.com?params=value
			timestamp := strconv.FormatInt(CreateTimestamp(), 10)  // int64 ==> string
			nonceStr := CreateNonceStr(NonceStrLength)
			signatureStr := "jsapi_ticket=" + instance.Ticket + "&noncestr=" + nonceStr + "&timestamp=" + timestamp + "&url=" + url
			signature := GetSha1(signatureStr)

			response.Rel = true						
			response.Result.TimeStamp = timestamp
			response.Result.NonceStr = nonceStr
			response.Result.Signature = signature
			response.Result.Appid = Appid
		}
		log.Println("*** 返回结果：", response)
		fmt.Println()			
	} else {
		response.Rel = false
		response.Msg = "wrong oauth, please sure that the params of the request key is correct."
		log.Println("请求失败，请求参数不正确")
		fmt.Println()			
	}
	fmt.Println()
	json.NewEncoder(w).Encode(response)
}
/* #endregion */



/* #region util */
// 检测变量是否为空
func IsEmpty(a interface{}) bool {
    v := reflect.ValueOf(a)
    if v.Kind() == reflect.Ptr {
        v=v.Elem()
    }
    return v.Interface() == reflect.Zero(v.Type()).Interface()
}

// 创建随机字符串
func CreateNonceStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	log.Println("随机字符串", string(result))
	return string(result)
}

// 创建时间戳
func CreateTimestamp() int64 {	
	t := time.Now().Unix()
	log.Println("当前时间时间戳", t)
	return t			
}

//SHA1加密
func GetSha1(data string) string {
	r := sha1.Sum([]byte(data))
	result := hex.EncodeToString(r[:])
	log.Println("sha1加密", result)
    return result
}
/* #endregion */