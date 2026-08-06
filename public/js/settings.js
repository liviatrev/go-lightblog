document.getElementById('uploadModeSelect').addEventListener('change', function () {
    var keyContainer = document.getElementById('imageKitKeyContainer');
    if (this.value === 'imagekit') {
        keyContainer.classList.remove('d-none');
        keyContainer.classList.add('d-block');
    } else {
        keyContainer.classList.remove('d-block');
        keyContainer.classList.add('d-none');
    }
});

document.getElementById('geminiSwitch').addEventListener('change', function () {
    var geminiContainer = document.getElementById('geminiAiContainer');
    var geminiHidden = document.getElementById('enable_gemini');
    if (this.checked) {
        geminiContainer.classList.remove('d-none');
        geminiContainer.classList.add('d-flex');
        geminiHidden.value = 'yes';
    } else {
        geminiContainer.classList.remove('d-flex');
        geminiContainer.classList.add('d-none');
        geminiHidden.value = 'no';
    }
});