
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

function triggerPageEvent(name, pageName, refresh=true) {
    if (name == null || name == "" || pageName == null || pageName == "") {
        return
    } 
    document.dispatchEvent(new CustomEvent(name, { detail: { page: pageName } }))
    ihui.trigger(name, pageName, null, refresh)
}

global.ihui = {}

function start() {

    var current_page

    var location = myScript.src
    if (window.location.protocol == "https:") {
        var protocol = "wss://"
        location = location.replace("https://", "")
    } else {
        var protocol = "ws://"
        location = location.replace("http://", "")
    }
    var ws = new WebSocket(protocol + location + '/ws');

    global.ihui.on = function (event, name, page, target, e) {
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
                $(e).find("input[name], textarea[name], select[name]").each(function(index,el){
                    var name = $(el).attr("name")
                    data[name] = $(el).val()
                })
                break;

            default:
                return
        }

        var msg = { name: name, page: page, id: id, target: target, data: data, refresh: true }
        // console.log(msg)
        ws.send(JSON.stringify(msg))
        event.preventDefault()
    }

    global.ihui.trigger = function (name, page, data, refresh=true) {
        var msg = { name: name, page: page, data: data, refresh: refresh}
        ws.send(JSON.stringify(msg))
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
        console.log(msg)

        switch (msg.Name) {
            case "init":
                localStorage.setItem("sessionId", msg.Id)
                break

            case "update":
                var el = $(msg.Target)
                updateHTML(el, msg.Data)
                break

            case "page":
                if (msg.Data.title && msg.Data.title != "") {
                    document.title = msg.Data.title
                }
                
                var page = $(msg.Target + " > #" + msg.Page)
                if (page.length > 0) {
                    updateHTML(page, msg.Data.html)
                    evt = "page-updated"
                } else {
                    if (msg.Data.replace)
                        $(msg.Target).html(msg.Data.html)
                    else
                        $(msg.Target).append(msg.Data.html)
                    evt = "page-created"
                }
                triggerPageEvent(evt, msg.Page, false)
                break

            case "hide-page":
                $(msg.Target + " > #" + msg.Page).css('display', 'none')
                triggerPageEvent("page-hidden", msg.Page, false)
                break

            case "show-page":
                $(msg.Target + " > #" + msg.Page).css('display', 'inline')
                triggerPageEvent("page-shown", msg.Page, false)
                break
                
            case "remove-page":
                $(msg.Target + " > #" + msg.Page).remove()
                triggerPageEvent("page-removed", msg.Page, false)
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

