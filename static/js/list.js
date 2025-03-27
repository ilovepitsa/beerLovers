function takePartClicked(elem) {
    var id = elem.getAttribute('data-id');
    var vote = 0;
    // console.log(elem.outerHTML)
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
    request.open('POST', '/events/take_part?id='+id+'&vote='+vote,true);
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
    request = new XMLHttpRequest();
    if (elem.classList.contains('btn-success')) {
        return;
    }
    request.open('GET', '/api/v1/user/balance?uid='+uid, true);
    

    request.onload = function () {
        
        var resp = JSON.parse(request.responseText);
        if(resp.error) {
            console.log("checkBalance server err:", resp.err);
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
        console.log("checkBalance  error", request.responseText);
    }
    request.send();
}


function addMoney(userId) {
    var form = new FormData(document.getElementById('refill'));
    var request = new XMLHttpRequest();
    request.open('POST', '/api/v1/user/balance', true)
    request.onload = function () {
        var resp = JSON.parse(request.responseText)
        if (resp.error) {
            console.log("addMoney err: ", resp.err)
            return
        }

        updateBalance(userId)
    }
    request.send(form)

}

function updateBalance(uid) {
    var request = new XMLHttpRequest();
    request.open('GET','/api/v1/user/balance', true);

    request.onload = function() {
        var resp = JSON.parse(request.responseText);
        if(resp.error) {
            console.log("update balance server err: ", resp.err);
            return;
        }
        var elem = document.getElementById('balance');
        elem.textContent = resp.body.balance.toFixed(2)  + " ₽";
    }
    request.onerror = function() {
        console.log("updateBalance  error", request.responseText);
    }

    request.send();
}

function changeLevel(uid, vote) {
    var request = new XMLHttpRequest();
    console.log(uid);
    request.open('PUT','/api/v1/user/changeLevel?uid='+uid+'&vote='+vote, true);

    request.send();
}


function deleteEvent(eid) {
    var request = new XMLHttpRequest();
    request.open('DELETE','/api/v1/event/delete?uid='+eid, true);


    request.send();
}

function deleteUser(uid) {
    var request = new XMLHttpRequest();
    request.open('DELETE','/api/v1/user/delete?uid='+uid, true);

    request.send();
}

function leaveFeedback(elem) {
    var feedbackForm = document.getElementById('feedbackForm');
    var formData = new FormData(feedbackForm); // Collect form data (including files)
    var eid = elem.dataset.event_id;
    
    var request = new XMLHttpRequest();
    request.open('POST', '/api/v1/event/review?eid=' + eid, true);

    request.onload = function () {
        $('#leaveFeedback').modal('hide');
        // Optional: Show success message or reset the form
        feedbackForm.reset();
    };

    request.send(formData); // Send FormData instead of plain text
}

function makeFavorite(elem) {
    var id = elem.dataset.id;
    var vote = 0;
    if (elem.classList.contains('btn-danger')) {
        elem.classList.remove('btn-danger');
        elem.classList.add('btn-primary');

        elem.textContent = 'Добавить в любимое';
        vote = -1;
    } else {
        elem.classList.add('btn-danger');
        elem.textContent = 'Любимое';
        vote = 1;
    }

    var request = new XMLHttpRequest();
    request.open('POST','/beer/make_favorite?id='+id+'&vote='+vote,true);
    request.send();
}


function getBeerFavoriteReport() {
    console.log('report!')
    fetch('/api/v1/reports/beer_favorite')
        .then(responce => responce.blob())
        .then(blob =>{
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;

            a.download = 'beer_favorite_report.csv'
            document.body.appendChild(a);
            a.click();
            setTimeout(() =>{
                document.body.removeChild(a);
                window.URL.revokeObjectURL(url);
            },0); 
        })
        .catch(error => console.error('Ошибка скачивания файла: ',error));
}