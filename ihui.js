morphdom = require('morphdom')

var ws

global.sendMsg = function (event, name, source, data) {
    if (event) {
        event.preventDefault()
    }
    var msg = JSON.stringify({ name: name, source: source, data: data })
    ws.send(msg)
}
global._s = function(event, source, data) {
    sendMsg(event, "action", source, data)
}


function updateHTML(page, html) {
    morphdom(page[0], html, {
        onBeforeElChildrenUpdated: function (fromEl, toEl) {
            if ($(toEl).hasClass('noupdate')) {
                return false
            }
            return true
        },
        childrenOnly: true
    })

//    page.trigger("ihui:display", page)
}

var currentPage = null


$(document).ready(function () {
    var protocol = "ws://"
    if (window.location.protocol == "https:") {
        protocol = "wss://"
    }
    addr = protocol + window.location.host + window.location.pathname + "ws"
    ws = new WebSocket(addr);

    ws.onerror = function(event) {
        console.log(event)
    }

    ws.onmessage = function (event) {
        var msg = JSON.parse(event.data);
        console.log(msg)
        var body = $(document.body)

        switch (msg.Name) { 
            case "new":
                document.title = msg.Data.title
                $("body > #main").html(msg.Data.html)
                break

            case "update":
                document.title = msg.Data.title
                updateHTML($("body > #main"), msg.Data.html)
                break

            case "script":
                // console.log(msg.Data)
                jQuery.globalEval(msg.Data)
                sendMsg(null, "script", "global", "ok")
                break
        }

    }

    ws.onclose = function (event) {
        console.log("Connection closed.")
        location.reload()
    }

});