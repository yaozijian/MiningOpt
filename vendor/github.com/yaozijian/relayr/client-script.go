/*
 *
 * This file contains client side Javascript that is rendered
 * when a client hits the RelayR route with a GET request.
 *
 */

package relayr

const connectionClassScript = `

var RelayRConnection = {};

RelayRConnection = (function() {

	var onReadyCalled = false;
	var route = "%v";
	var web;

	var transport = {
		websocket: {
			connect: function(onDataCallback) {

				var s = this;

				s.socket = new WebSocket(
					"ws://" + window.location.host + route
					+ "/ws?connectionId=" + transport.ConnectionId
				);

				// 套接字关闭时稍后重新开始协商
				s.socket.onclose = function(evt) {
					setTimeout(function() { web.negotiate(); }, 2000);
				};

				// 出错时重新连接
				s.socket.onerror = function(evt) {
					setTimeout(function(){s.connect(onDataCallback);},0);
				};

				s.socket.onopen = function(evt) {
					if (!onReadyCalled) {
						RelayRConnection.onReadyCallback();
						onReadyCalled = true;
					}
				};

				// 收到数据时调用回调函数
				s.socket.onmessage = function(evt) {
					onDataCallback(evt.data);
				};
			},

			send: function(data) {
				this.socket.send(data);
			}
		},
		longpoll: {
			connect: function(onDataCallback) {

				if (!onReadyCalled) {
					RelayRConnection.onReadyCallback();
					onReadyCalled = true;
				}

				var retry;

				retry = function() {
					web.doAjaxGetJson(
						route + '/longpoll?' + this.reqarglist(),
						function(data) {
							if (data.responseText) {
								var reconn = JSON.parse(data.responseText);
								if (reconn.Z) {
									web.negotiate();
								} else {
									onDataCallback(data);
									retry();
								}
							} else {
								web.negotiate();
							}
					});
				};

				retry();
			},

			send: function(data) {
				web.doAjaxPost(route + '/call?' + this.reqarglist(),data, null,null);
			},

			reqarglist: function(){
				var arglist = 'connectionId=' + transport.ConnectionId;
				arglist = arglist + '&_=' + new Date().getTime();
				return arglist;
			}
		}
	};

	web = (function() {
		return {
			/* 协商通信方法 */
			negotiate: function() {

				var ptr = this;

				var requrl = route + "/negotiate?_=" + new Date().getTime();

				var transportType = ptr.getTransportType();
				var handshake = JSON.stringify({t: transportType});

				ptr.doAjaxPost(
					requrl,handshake,
					/* 协商完成的处理函数 */
					function(result) {

						var obj = JSON.parse(result.responseText);
						/* 用ConnectionId 标识一个客户端(浏览器) */
						transport.ConnectionId = obj.ConnectionID;

						setTimeout(function() {
							/*
								协商通信方法完成,用指定的通信方法发起连接.
								参数为收到数据的回调函数.
								每次收到的数据是一个Web服务器调用客户端(JavaScript)函数的请求.
							*/
							transport[transportType].connect(function(data) {

								var callreq;

								// 调用请求
								if (data.responseText && data.status && data.responseXML) {
									callreq = JSON.parse(data.responseText);
								} else {
									if (data.responseText == "") return;
									callreq = JSON.parse(data);
								}

								// 参数列表
								var arglist = [];
								for (var i = 0; i < callreq.A.length; i++) {
									arglist.push(callreq.A[i]);
								}

								// 调用客户端(JavaScript)对象的指定方法
								var jsobj = RelayR[callreq.R].client;
								jsobj[callreq.M].apply(jsobj || window, arglist);
							});
						}, 0);
					},
					/* 如果失败,则2秒后重试 */
					function(result) {
						setTimeout(function() {ptr.negotiate();}, 2000);
					}
				);
			},

			/* 执行 Ajax POST 操作*/
			doAjaxPost: function(url, data, readyCallback, errCallback) {

				var ajaxobj = this.newAjaxObj();

				if (readyCallback != null){
					ajaxobj.onreadystatechange = function() {
						if ((ajaxobj.readyState === 4) && (ajaxobj.status === 200)){
							readyCallback(ajaxobj);
						}
					};
				}

				if (errCallback != null){
					ajaxobj.onerror = function() {errCallback(ajaxobj);};
				}

				ajaxobj.open('POST', url, true);
				ajaxobj.setRequestHeader("Content-type", "application/json");
				ajaxobj.send(data);

				window.onbeforeunload = function() {
					delete ajaxobj;
					ajaxobj = null;
				};
			},

			/* 执行 Ajax GET 操作*/
			doAjaxGet: function(url,readyCallback){
				this.ajaxGetOp(url,readyCallback,false);
			},

			doAjaxGetJson: function(url,readyCallback){
				this.ajaxGetOp(url,readyCallback,true);
			},

			ajaxGetOp: function(url, readyCallback, usejson) {

				var ajaxobj = this.newAjaxObj();

				ajaxobj.onreadystatechange = function() {
					if (ajaxobj.readyState === 4 && ajaxobj.status === 200) {
						readyCallback(ajaxobj);
					}
				};

				ajaxobj.open('GET', url, true);

				if (usejson){
					ajaxobj.setRequestHeader("Content-type", "application/json");
				}

				ajaxobj.send();
			},

			newAjaxObj: function() {

				var ajaxobj;

				if (window.XMLHttpRequest) {
					ajaxobj = new XMLHttpRequest();
				}else {
					ajaxobj = new ActiveXObject("Microsoft.XMLHTTP");
				}

				return ajaxobj;
			},

			getTransportType: function() {
				if (window.WebSocket) {
					return "websocket";
				} else {
					return "longpoll";
				}
			}
		};
	})();

	return {
		ready: function(onReadyCallback) {
			RelayRConnection.onReadyCallback = onReadyCallback;
			web.negotiate();
		},
		callServer: function(r, f, a) {
			transport[web.getTransportType()].send(
				JSON.stringify({
					S: true,
					C: transport.ConnectionId,
					R: r,
					M: f,
					A: a
				})
			);
		}
	};
})();

`

const relayClassBegin = `

var RelayR = (function() {
	return {

`

const relayBegin = `

%v: {

	client: {},

	server: {

`

// {0} == function name
// {1} == Relay name
// {2} == function name
const relayMethod = `

%v: function() {
	RelayRConnection.callServer('%v', '%v', Array.prototype.slice.call(arguments));
},

`

const relayEnd = `

},

},

`

const relayClassEnd = `

	};
})();

`
