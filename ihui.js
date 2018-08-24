
var morphdom = require('morphdom')

var ws

global.trigger = function (event, name, source, target, data) {
    if (event) {
        event.preventDefault()
    }
    var msg = JSON.stringify({ name: name, source: source, target: target, data: data })
    ws.send(msg)
}
global.sendMsg = global.trigger


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

}


$(document).ready(function () {
    var protocol = "ws://"
    if (window.location.protocol == "https:") {
        protocol = "wss://"
    }
    addr = protocol + window.location.host + "{{.Path}}/ws"
    ws = new WebSocket(addr);

    ws.onerror = function(event) {
        console.log(event)
    }

    ws.onmessage = function (event) {
        var msg = JSON.parse(event.data);
        console.log(msg)
        var body = $(document.body)

        if (msg.Data.title && msg.Data.title != "") {
            document.title = msg.Data.title
        }

        switch (msg.Name) {
            case "new":
                $("body > #main").html(msg.Data.html)
                $(document).trigger("page-new")
                trigger(null, "load", "page", "page", null)
                break

            case "update":
                updateHTML($("body > #main"), msg.Data.html)
                $(document).trigger("page-update")
                break

            case "script":
                // console.log(msg.Data)
                jQuery.globalEval(msg.Data)
                trigger(null, "script", "global", "ok")
                break
        }

    }

    ws.onclose = function (event) {
        console.log("Connection closed.")
        location.reload()
    }

});