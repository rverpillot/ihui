morphdom = require('morphdom')

var ws

global.sendMsg = function (name, source, data) {
    ws.send(JSON.stringify({ name: name, source: source, data: data }))
}

function handleEvents(el) {
    var id = $(el).attr("id")
    if (!id) return
    var action = $(el).attr("data-action")

    switch (action) {
        case "click":
            $(el).one("click", { id: id }, function (event) {
                console.log(event)
                event.preventDefault()
                sendMsg("click", event.data.id, null)
            })
            break

        case "check":
            $(el).on("change", { id: id }, function (event) {
                event.preventDefault();
                sendMsg("check", event.data.id, $(this).prop("checked"))
            })
            break

        case "change":
            $(el).on("change", { id: id, }, function (event) {
                event.preventDefault();
                sendMsg("change", event.data.id, $(this).val())
            })
            break

        case "form":
            $(el).find("input[name], textarea[name], select[name]").on("change", { id: id, }, function (event) {
                event.preventDefault();
                sendMsg("change", event.data.id, { name: $(this).attr("name"), val: $(this).val() })
            })
            break

        case "input":
            $(el).on("input", { id: id, }, function (event) {
                event.preventDefault();
                sendMsg("change", event.data.id, $(this).val())
            })
            break

        case "select":
            break

        case "submit":
            $(el).on("submit", { id: id, }, function (event) {
                event.preventDefault();
                sendMsg("form", event.data.id, $(this).serializeObject())
            })
            break
    }


}

function updateHTML(page, html) {
    page.find("[data-action]").off()
    page.find("[data-action] input").off()
    page.find("[data-action] textarea").off()
    page.find("[data-action] select").off()

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

    page.find("[data-action]").each(function (index, el) { handleEvents(el) })

    page.trigger("ihui:display", page)
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