
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

        if (msg.Data.title && msg.Data.title != "") {
            document.title = msg.Data.title
        }

        switch (msg.Name) {
            case "page":
                if ($(".page").is("#" + msg.Data.name)) {
                    updateHTML($("#"+msg.Data.name), msg.Data.html)
                } else {
                    $("#pages").append(msg.Data.html)
                }
                showPage(msg.Data.name)
                $(document).trigger("page-"+msg.Name, msg.Data.name)
                trigger(null, msg.Name, "page", "page", null)
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

