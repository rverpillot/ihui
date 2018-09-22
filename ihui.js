
var morphdom = require("morphdom")
var $ = require("jquery")
// var $ = require("zepto")

var scripts = document.getElementsByTagName('script')
var myScript = scripts[scripts.length - 1]

function updateHTML(el, html) {
    if (!el) {
        return
    }
    morphdom(el, html, {
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
    document.querySelector(".page#" + name).style.display = ''
}

global.ihui = {}

global.ihui.start = function () {

    if ($("#pages").length == 0) {
        $("body").prepend('<div id="pages"></div>')
    }
    var current_page

    var location = myScript.src.replace("/js/ihui.js", "")
    if (window.location.protocol == "https:") {
        var protocol = "wss://"
        location = location.replace("https://", "")
    } else {
        var protocol = "ws://"
        location = location.replace("http://", "")
    }
    var ws = new WebSocket(protocol + location + "/ws");

    global.ihui.on = function (event, name, target, e) {
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
                var form = $(e).serializeArray()
                var data = {}
                for (var i = 0; i < form.length; i++) {
                    data[form[i].name] = form[i].value
                }
                    break;

                default:
                    return
            }
            
            var msg = { name: name, id: id, target: target, data: data }
            // console.log(msg)
            ws.send(JSON.stringify(msg))
            history.pushState(msg, "")
            event.preventDefault()
    }

    global.ihui.trigger = function (name, target, data) {
            var msg = { name: name, target: target, data: data }
            ws.send(JSON.stringify(msg))
        }


    window.onpopstate = function (event) {
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
            case "update":
                var el = document.querySelector(msg.Target)
                updateHTML(el, msg.Data)
                break

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


