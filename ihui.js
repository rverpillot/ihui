
var morphdom = require("morphdom")
var $ = require("cash-dom")

function updateHTML(el, html, childrenOnly = false) {
    for (var i = 0; i < el.length; i++) {
        morphdom(el[i], html, {
            onBeforeElUpdated: function (fromEl, toEl) {
                if (fromEl.isEqualNode(toEl) || toEl.classList.contains('noupdate')) {
                    return false
                }
                return true
            },
            childrenOnly: childrenOnly
        })
    }
}

global.ihui = {}

function start() {
    var url = location.href
    if (window.location.protocol == "https:") {
        url = url.replace("https://", "wss://")
    } else {
        url = url.replace("http://", "ws://")
    }
    var ws = new WebSocket(url + 'ihui.js/ws');

    var last_page;
    var quit = false;

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
                updateHTML(el, msg.Data, true)
                break

            case "element":
                if (msg.Data.title && msg.Data.title != "") {
                    document.title = msg.Data.title
                }
                if (msg.Data.page) {
                    if (last_page != msg.Element) {
                        last_page = msg.Element
                        window.scrollTo(0, 0)
                    }
                    $(".ihui-page").not("#" + msg.Element).css('display', 'none') // display only the current page
                }
                var element = $("#" + msg.Element)
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
                var element = $(msg.Target + " > #" + msg.Element)
                if (msg.Data.page) {
                    element.remove()
                } else {
                    element.css('display', 'none').html("")
                }
                break

            case "script":
                if (msg.Data != "") {
                    eval(msg.Data)
                }
                break

            case "quit":
                $("body").html("")
                quit = true
                ws.close()
                break
        }

    }

    ws.onclose = function (event) {
        console.log("Connection closed.")
        if (!quit)
            window.location.reload()
    }
}

window.addEventListener("load", function (event) {
    start()
})

