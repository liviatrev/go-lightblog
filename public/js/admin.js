document.addEventListener('click', function (e) {
    var btn = e.target.closest('[data-copy-api-key]');
    if (!btn) return;

    var key = btn.dataset.copyApiKey;
    navigator.clipboard.writeText(key).then(function () {
        alert('API Key berhasil disalin ke clipboard!');
    }).catch(function (err) {
        alert('Gagal menyalin API Key.');
        console.error(err);
    });
});

document.addEventListener('submit', function (e) {
    var form = e.target.closest('[data-confirm]');
    if (!form) return;

    var message = form.dataset.confirm;
    if (!confirm(message)) {
        e.preventDefault();
    }
});

document.addEventListener('click', function (e) {
    var link = e.target.closest('[data-confirm]');
    if (!link || link.tagName === 'FORM') return;

    var message = link.dataset.confirm;
    if (!confirm(message)) {
        e.preventDefault();
    }
});
