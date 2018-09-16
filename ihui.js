
var morphdom = require('morphdom')


function updateHTML(page, html) {
    morphdom(page, html, {
        onBeforeElUpdated: function (fromEl, toEl) {
            if (toEl.classList.contains('noupdate')) {
                return false
            }
            return true
        },
        childrenOnly: true
    })

}

function showPage(name) {
    var pages = document.querySelectorAll(".page")
    for (var i = 0; i < pages.length; i++) {
        pages[i].style.display = 'none'
    }
    document.querySelector(".page#"+name).style.display = ''
}

function start() {
    var protocol = "ws://"
    if (window.location.protocol == "https:") {
        protocol = "wss://"
    }
    var current_page

    var addr = protocol + window.location.host + "{{.Path}}/ws"
    var ws = new WebSocket(addr);

    global.ihui = {
        on: function (event, name, target, e) {
            var id = $(e).attr("data-id") || $(e).attr("id") || ""

            switch (name) {
                case "click":
                    var data = $(e).attr("data-value") || $(e).attr("data-id") || $(e).attr("id") || ""  
                    break;
            
                case "check":
                    var data = $(e).prop("checked")
                    break;

                case "change":
                    var nameAttr = $(e).attr("name")
                    var value = $(e).val()
                    if (nameAttr) {
                        var data = {name: nameAttr, val: value}
                    } else {
                        var data = value
                    }
                    break;

                case "input":
                    var data = $(e).val()
                    break;

                case "submit":
                    var data = $(e).serializeObject()
                    break;

                default:
                    return
            }
            
            var msg = { name: name, id: id, target: target, data: data }
            ws.send(JSON.stringify(msg))
            history.pushState(msg, "")
            event.preventDefault()
        },

        trigger: function (name, target, data) {
            var msg = { name: name, target: target, data: data }
            ws.send(JSON.stringify(msg))
        }
    }

    window.onpopstate = function(event) {
        var msg = event.state
        if (!msg) {
            location.reload()
            return
        } 
        // console.log(msg)
        ihui.trigger(msg.name, msg.target, msg.data)
    }

    ws.onerror = function (event) {
        console.log(event)
    }

    ws.onmessage = function (event) {
        var msg = JSON.parse(event.data);
        // console.log(msg)

        switch (msg.Name) {
            case "page":
                if (msg.Data.title && msg.Data.title != "") {
                    document.title = msg.Data.title
                }

                var pageName = msg.Data.name
                if (pageName != current_page) {
                    current_page = pageName
                    window.scrollTo(0, 0)
                }

                var page = document.querySelector("#"+pageName)

                if (page) {
                    updateHTML(page, msg.Data.html)
                    evt = "update"
                } else {
                    $("#pages").append(msg.Data.html)
                    evt = "create"
                }
                showPage(pageName)
                $(document).trigger("page-" + evt, { page: pageName })
                ihui.trigger(evt, "page", pageName)
                break

            case "remove":
                var pageName = msg.Target
                var page = document.querySelector("#" + pageName)
                page.parentNode.removeChild(page)
                $(document).trigger("page-remove", { page: pageName })
                ihui.trigger("remove", "page", pageName)
                break

            case "script":
                eval(msg.Data)
                break
        }

    }

    ws.onclose = function (event) {
        console.log("Connection closed.")
        location.reload()
    }
}

window.addEventListener("load", function(event){
    start()
})

