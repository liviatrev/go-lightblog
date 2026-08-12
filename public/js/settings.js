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

document.getElementById('cloudflareSwitch').addEventListener('change', function () {
    var cloudflareContainer = document.getElementById('cloudflareContainer');
    var cloudflareHidden = document.getElementById('enable_cloudflare');
    if (this.checked) {
        cloudflareContainer.classList.remove('d-none');
        cloudflareContainer.classList.add('d-flex');
        cloudflareHidden.value = 'yes';
    } else {
        cloudflareContainer.classList.remove('d-flex');
        cloudflareContainer.classList.add('d-none');
        cloudflareHidden.value = 'no';
    }
});

document.getElementById('indexnowSwitch').addEventListener('change', function () {
    var indexnowContainer = document.getElementById('indexnowContainer');
    var indexnowHidden = document.getElementById('indexnow');
    if (this.checked) {
        indexnowContainer.classList.remove('d-none');
        indexnowContainer.classList.add('d-flex');
        indexnowHidden.value = 'yes';
    } else {
        indexnowContainer.classList.remove('d-flex');
        indexnowContainer.classList.add('d-none');
        indexnowHidden.value = 'no';
    }
});

// Theme Picker: click a swatch to select the theme
document.querySelectorAll('.theme-swatch').forEach(function (swatch) {
    swatch.addEventListener('click', function () {
        var theme = this.getAttribute('data-theme');
        var select = document.getElementById('public_theme');

        if (select) {
            select.value = theme;
            // Remove previous outline on all swatches
            document.querySelectorAll('.theme-swatch').forEach(function (s) {
                s.style.outline = '';
            });
            // Highlight the selected swatch
            this.style.outline = '2px solid #0d6efd';
            this.style.outlineOffset = '1px';
        }
    });
});