function takePartClicked(elem) {
    var id = elem.getAttribute('data-id');
    var takePart = 0;
    if(elem.classList.contains('btn-success')) {
        elem.classList.remove('btn-success');
        vote = -1;
    } else {
        elem.classList.add('btn-success');
        vote = 1;
    }

    var request = new XMLHttpRequest();
    request.open('POST', '/events/take_part?id='+id+'&vote='+vote,true);
    // request.onload = function() {
    //     var resp = JSON.parse(request.responseText)
    //     if(resp.err) {
    //         console.log("take part server err: ", resp.err)
    //         return;
    //     }
    // }

    request.send();
    
}



function uploadPhoto(uid) {
    var form = new FormData(document.getElementById('add_beer'))
    var request = new XMLHttpRequest();
    request.open('POST', '/beer/create', true);
    request.onload = function() {
        var resp = JSON.parse(request.responseText);
        if(resp.error) {
            console.log("rateComment server err:", resp.err);
            return;
        }
        
    };
    request.send(form);
}



function checkBalance(uid, elem) {
    request = new XMLHttpRequest();
    request.Open('GET', 'user/getBalance?uid='+uid, true)


    request.onload = function () {
        var resp = JSON.parse(request.responseText)
        if(resp.error) {
            console.log("renderPhotos server err:", resp.err);
            return;
        }

        if (resp.body.balance )

    }
    
}