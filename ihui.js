
var morphdom = require('morphdom')

var ws

global.trigger = function (event, name, source, target, data) {
    if (event) {
        event.preventDefault()
    }
    var msg = JSON.stringify({ name: name, source: source, target: target, data: data })
    ws.send(msg)
}

function updateHTML(page, html) {
    morphdom(page[0], html, {
        onBeforeElUpdated: function (fromEl, toEl) {
            if ($(toEl).hasClass('noupdate')) {
                return false
            }
            return true
        },
        childrenOnly: true
    })
    
}

function showPage(name) {
    $(".page").each(function(i,e){
        var page = $(e)
        if (page.attr("id") == name) {
            page.show()
        } else {
            page.hide()
        }
    })
}

function start() {
    var protocol = "ws://"
    if (window.location.protocol == "https:") {
        protocol = "wss://"
    }
    var addr = protocol + window.location.host + "{{.Path}}/ws"
    ws = new WebSocket(addr);

    ws.onerror = function(event) {
        console.log(event)
    }

    ws.onmessage = function (event) {
        var msg = JSON.parse(event.data);
        console.log(msg)
        var body = $(document.body)

        switch (msg.Name) {
            case "page":
                if (msg.Data.title && msg.Data.title != "") {
                    document.title = msg.Data.title
                }

                if ($(".page").is("#" + msg.Data.name)) {
                    updateHTML($("#"+msg.Data.name), msg.Data.html)
                    evt = "update"
                } else {
                    $("#pages").append(msg.Data.html)
                    evt = "create"
                }
                showPage(msg.Data.name)
                $(document).trigger("page-"+evt, msg.Data.name)
                trigger(null, evt, msg.Data.name, "page", null)
                break

            case "remove":
                $("#"+msg.Target).remove()
                $(document).trigger("page-remove", msg.Target)
                trigger(null, "remove", msg.Target, "page", null)
                break

            case "script":
                jQuery.globalEval(msg.Data)
                break
        }

    }

    ws.onclose = function (event) {
        console.log("Connection closed.")
        location.reload()
    }
}

$(window).on("load", function () { 
    start(); 
})

