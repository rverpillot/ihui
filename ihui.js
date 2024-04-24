
var morphdom = require("morphdom")
var $ = require("cash-dom")

var scripts = $('script')
var myScript = scripts.last()[0]

function updateHTML(el, html) {
    for (var i = 0; i < el.length; i++) {
        morphdom(el[i], html, {
            onBeforeElUpdated: function (fromEl, toEl) {
                if (toEl.classList.contains('noupdate')) {
                    return false
                }
                return true
            },
            childrenOnly: false
        })
    }
}

global.ihui = {}

function start() {

    var location = myScript.src
    if (window.location.protocol == "https:") {
        var protocol = "wss://"
        location = location.replace("https://", "")
    } else {
        var protocol = "ws://"
        location = location.replace("http://", "")
    }
    var ws = new WebSocket(protocol + location + '/ws');

    function sendEvent(name, elementName, id, target, data, refresh = true) {
        var msg = { name: name, element: elementName, id: id, target: target, data: data, refresh: refresh }
        // console.log("Send:", msg)
        ws.send(JSON.stringify(msg))
    }

    function triggerElementEvent(name, elementName, refresh = true) {
        document.dispatchEvent(new CustomEvent(name, { detail: { element: elementName } }))
        sendEvent(name, elementName, null, null, null, refresh)
    }

    global.ihui.trigger = function (name, element, data, refresh = true) {
        sendEvent(name, element, null, null, data, refresh)
    }

    global.ihui.on = function (event, name, element, target, e) {
        var id = $(e).attr("data-id") || $(e).attr("id") || ""

        switch (name) {
            case "click":
                var win = $(e).attr("target")
                var data = $(e).attr("data-value") || $(e).attr("data-id") || $(e).attr("id") || ""
                if (win) {
                    data = { target: win, val: data }
                    window.open($(e).attr("href") || "", win)
                }
                break;

            case "check":
                var data = $(e).prop("checked")
                break;

            case "change":
                var data = $(e).val()
                break;

            case "form":
                var data = { name: $(e).attr("name"), val: $(e).val() }
                break;

            case "input":
                var data = $(e).val()
                break;

            case "submit":
                var data = {}
                $(e).find("input[name], textarea[name], select[name]").each(function (index, el) {
                    var name = $(el).attr("name")
                    data[name] = $(el).val()
                })
                break;

            default:
                return
        }

        sendEvent(name, element, id, target, data)
        event.preventDefault()
    }

    // window.onpopstate = function (event) {
    //     var msg = event.state
    //     if (!msg) {
    //         window.location.reload()
    //         return
    //     }
    //     console.log(msg)
    // }

    ws.onerror = function (event) {
        console.log(event)
    }

    ws.onopen = function (event) {
        var sessionId = localStorage.getItem("sessionId") || ""
        ws.send(JSON.stringify({ Name: "connect", Id: sessionId }))
    }

    ws.onmessage = function (event) {
        var msg = JSON.parse(event.data);
        // console.log("Receive:", msg)

        switch (msg.Name) {
            case "init":
                localStorage.setItem("sessionId", msg.Id)
                break

            case "update":
                var el = $(msg.Target)
                updateHTML(el, msg.Data)
                break

            case "element":
                if (msg.Data.title && msg.Data.title != "") {
                    document.title = msg.Data.title
                }
                if (msg.Data.page) {
                    $(".ihui-page").not("#" + msg.Element).css('display', 'none') // display only the current page
                }
                var element = $(msg.Target + " > #" + msg.Element)
                if (element.length > 0) {
                    updateHTML(element, msg.Data.html)
                    evt = "element-updated"
                } else {
                    if (msg.Data.replace)
                        $(msg.Target).html(msg.Data.html)
                    else
                        $(msg.Target).append(msg.Data.html)
                    evt = "element-created"
                }
                triggerElementEvent(evt, msg.Element, false)
                break

            case "hide":
                $(msg.Target + " > #" + msg.Element).css('display', 'none')
                break

            case "show":
                $(msg.Target + " > #" + msg.Element).css('display', 'inline')
                break

            case "remove":
                $(msg.Target + " > #" + msg.Element).remove()
                break

            case "script":
                eval(msg.Data)
                break
        }

    }

    ws.onclose = function (event) {
        console.log("Connection closed.")
        window.location.reload()
    }
}

window.addEventListener("load", function (event) {
    start()
})

