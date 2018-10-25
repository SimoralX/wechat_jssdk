# wechat_jssdk
The golang code which can get [wechat jssdk](https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421141115).

## Router
Thanks to golang package [mux](https://github.com/gorilla/mux)

| router | method | remarks |
| -- | :--: | -- |
| /jssdk/gettoken  | get | get token |
| /jssdk/getticket | get | get ticket |
| /jssdk/getconfig | get | get signature |
| /static/         | get | static server, you can change the static file path |

## How to use?
Change the property like this:
```
// appid
const Appid = "**********************************"
// secret
const AppSecret = "********************************"
```
which in jssdk.go.

