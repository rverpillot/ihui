
var morphdom = require('morphdom')


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
    var current_page

    var addr = protocol + window.location.host + "{{.Path}}/ws"
    var ws = new WebSocket(addr);

    global.trigger = function (event, name, id, target, data) {
        if (event) {
            event.preventDefault()
        }
        var msg = JSON.stringify({ name: name, id: id, target: target, data: data })
        ws.send(msg)
    }


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

                var page = msg.Data.name
                if (page != current_page) {
                    current_page = page
                    window.scrollTo(0, 0)
                }

                if ($(".page").is("#" + page)) {
                    updateHTML($("#"+page), msg.Data.html)
                    evt = "update"
                } else {
                    $("#pages").append(msg.Data.html)
                    evt = "create"
                }
                showPage(page)
                $(document).trigger("page-"+evt, {page: page})
                trigger(null, evt, page, "page", null)
                break

            case "remove":
                $("#"+msg.Target).remove()
                $(document).trigger("page-remove", {page: msg.Target})
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

