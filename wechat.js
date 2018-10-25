// 微信分享,引入weixin-js-sdk包
import wx from 'weixin-js-sdk'
// 引入axios包
import axios from 'axios'




export function share (reqUrl, obj) {        
    let splitUrl = reqUrl.split('#')[0];

    // 二次分享处理尝试，失败，不处理成功。
    // let re = '?from=singlemessage';
    // if (splitUrl.lastIndexOf(re) >= 0) {
    //     splitUrl = splitUrl.replace(re, '');        
    // }
    
    axios.get("******** url ********", { params: {                  
        url: splitUrl
    }}).then((res) => {
        console.log("share()：", res);
        if (res.data.rel) { 
            init(res.data.result, splitUrl, obj); 
        }
    }).catch((err) => {
        console.error(err);
    })
}


export function init (res, url, obj) {    
    let config = {    
        debug: false,  // 开启调试模式,调用的所有api的返回值会在客户端alert出来，若要查看传入的参数，可以在pc端打开，参数信息会通过log打出，仅在pc端时才会打印。
        appId: res.appId,  // 必填，公众号的唯一标识
        timestamp: res.timestamp,  // 必填，生成签名的时间戳
        nonceStr: res.nonceStr,  // 必填，生成签名的随机串
        signature: res.signature,  // 必填，签名
        jsApiList: [  // 调用请求的接口在config.js文件内
             'onMenuShareTimeline',  // 分享到朋友圈接口             
             'onMenuShareAppMessage',  // 分享给朋友                                         
             'getLocation',  // 获取地理位置接口                                        
             'chooseWXPay'  // 发起一个微信支付请求
        ]  
    }
    console.log("init():[config]", config);
    
    /** 
     * [参考网址](https://mp.weixin.qq.com/wiki?action=doc&id=mp1421141115&t=0.04168839253989076&token=&lang=zh_CN#63)
     */
    wx.config(config);
    wx.ready (() => {                                      
        updateAppMessageShareData(url, obj);
        updateTimelineShareData(url, obj);                  
    });
    wx.error(function(err){
        // config信息验证失败会执行error函数，如签名过期导致验证失败，具体错误信息可以打开config的debug模式查看，也可以在返回的res参数中查看，对于SPA可以在这里更新签名。
        console.log("init(): config获取失败", err);                
    });
}


/** 基础设置 */
export function setOptions (url, obj) {
    
    let re = /&isappinstalled=0/gi;
    url = url.replace(re, '');
    let re1 = /\?from=singlemessage/gi;
    url = url.replace(re1, '');

    // let re = '?from=singlemessage';
    // if (url.lastIndexOf(re) >= 0) {
    //     url = url.replace(re, '');        
    // }    

    let options = {};
    if (obj && obj.activityId) {
        options.title = obj.title;
        options.desc = title;
        options.imgUrl = obj.img; 
        options.url = url + '#/activity-detail?activityId=' + obj.activityId;      
    } else {
        options.title = title;
        options.desc = desc;
        let baseUrl = url.split("index.html")[0];  // 部署到tomcat必须要加上
        options.imgUrl = baseUrl + 'static/image/shareLogo.png'
        options.url = url;
    }
    console.log('setOptions(), 基础设置', options);
    return options;
}


/** 分享给朋友 */
export function updateAppMessageShareData (url, obj) {            
    let options = setOptions(url, obj);

    console.log("updateAppMessageShareData()：[options] ", options);

    // 需在用户可能点击分享按钮前就先调用
    wx.onMenuShareAppMessage({ 
        title: options.title, // 分享标题
        desc: options.desc, // 分享描述
        link: options.url, // 分享链接，该链接域名或路径必须与当前页面对应的公众号JS安全域名一致
        imgUrl: options.imgUrl, // 分享图标
        // 用户确认分享后执行的回调函数
        success: function () {
            this.Debug('wechat.js/updateAppMessageShareData()', '分享成功');
        },
        // 用户取消分享后执行的回调函数
        cancel: function () {
            this.Debug('wechat.js/updateAppMessageShareData()', '取消分享');
        }
    });
}


/** 分享给朋友圈 */
export function updateTimelineShareData (url, obj) {
    let options = setOptions(url, obj);

    //需在用户可能点击分享按钮前就先调用
    wx.onMenuShareTimeline({ 
        title: options.title, // 分享标题
        link: options.url, // 分享链接，该链接域名或路径必须与当前页面对应的公众号JS安全域名一致
        imgUrl: options.imgUrl, // 分享图标
        success: function () {                        
            console.log('wechat.js/updateTimelineShareData()', '分享朋友圈成功');
            let toast = this.$createToast({
                time: 1000,
                txt: '分享朋友圈成功',
                type: 'correct' 
            })
            toast.show();
        },
        cancel: function () {            
            console.log('wechat.js/updateTimelineShareData()', '分享朋友圈失败');
            let toast = this.$createToast({
                time: 1000,
                txt: '分享朋友圈失败',
                type: 'error' 
            })
            toast.show();
        }
    });                        
}
