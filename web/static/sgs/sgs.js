function toastsMsg(msg) {
    document.getElementById("toast-body").innerHTML = msg;
    const toastLiveExample = document.getElementById('liveToast')
    const toastBootstrap = bootstrap.Toast.getOrCreateInstance(toastLiveExample)
    toastBootstrap.show()
}

function onUpload(api, domId) {
    let id = '#' + domId;
    let formData = new FormData();
    let fileCount = $(id)[0].files.length;
    for (let i = 0; i < fileCount; i++) {
        formData.append('files', $(id)[0].files[i]);
    }
    $.ajax({
        url: api,
        type: 'POST',
        async: false,
        data: formData,
        cache: false,
        contentType: false,
        processData: false,
        success: function (data) {
            alert(data);
        }
    });
}