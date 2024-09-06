function takePartClicked(cost, elem) {
    var id = elem.getAttribute('data-id');
    var vote = 0;
    console.log(elem.outerHTML)
    if (elem.classList.contains('disabled')) {
        return;
    }
    if(elem.classList.contains('btn-success')) {
        elem.classList.remove('btn-success');
        vote = -1;
    } else {
        elem.classList.add('btn-success');
        vote = 1;
    }

    var request = new XMLHttpRequest();
    request.open('POST', '/events/take_part?id='+id+'&vote='+vote+'&cost='+cost,true);
    request.send();
    
}



function uploadPhoto(uid) {
    var form = new FormData(document.getElementById('add_beer'));
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



function checkBalance(uid, cost,  elem) {
    console.log("uid: " + uid)
    request = new XMLHttpRequest();
    if (elem.classList.contains('btn-success')) {
        return;
    }
    request.open('GET', '/api/v1/user/balance?uid='+uid, true);
    

    request.onload = function () {
        
        var resp = JSON.parse(request.responseText);
        if(resp.error) {
            console.log("renderPhotos server err:", resp.err);
            return;
        }

        if (resp.body.balance < cost ) {
            console.log("not enought money");
            elem.classList.add('btn-warning');
            elem.classList.add('disabled');
            return
        }
        elem.classList.remove('btn-warning');
        elem.classList.remove('disabled');

    }

    request.onerror = function() {
        console.log("checkBalance  error", request.responseText)
    }
    request.send();
}


function addMoney(amount, userId) {

}