morphdom = require('morphdom')

var ws

global.sendMsg = function (name, source, data) {
    var msg = JSON.stringify({ name: name, source: source, data: data })
    ws.send(msg)
}


function updateHTML(page, html) {
    morphdom(page[0], html, {
        onBeforeElChildrenUpdated: function (fromEl, toEl) {
            if ($(toEl).hasClass('noupdate')) {
                return false
            }
            return true
        },
        childrenOnly: true,
        
        // onNodeAdded: function(el) {
        //     if ($(el).attr("data-action")) {
        //         handleEvents(el)
        //     }
        // },
        // onElUpdated: function(el) {
        //     if ($(el).attr("data-action")) {
        //         handleEvents(el)
        //     }
        // }
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

    //    ws.onerror = function(event) {}

    ws.onmessage = function (event) {
        var msg = JSON.parse(event.data);
        console.log(msg)
        var body = $(document.body)

        switch (msg.Name) {
            case "update":
                document.title = msg.Data.title
                updateHTML(body, msg.Data.html)
                if (msg.Source != currentPage) {
                    window.scrollTo(0, 0)
                    currentPage = msg.Source
                }
                break

            case "script":
                // console.log(msg.Data)
                jQuery.globalEval(msg.Data)
                break
        }

    }

    ws.onclose = function (event) {
        alert("Connection closed.")
        location.reload()
    }

});