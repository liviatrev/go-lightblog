(function () {
    var rawContentElement = document.getElementById('raw-content');
    var postBodyElement = document.getElementById('post-body');

    if (rawContentElement && postBodyElement) {
        postBodyElement.innerHTML = rawContentElement.textContent;
    }

    var remarkEl = document.getElementById('remark42');
    if (!remarkEl) return;

    var remarkConfig = {
        host: remarkEl.dataset.remarkHost,
        site_id: remarkEl.dataset.remarkSiteId,
        components: ['embed'],
        max_shown_comments: 15,
        theme: 'light',
        page_title: remarkEl.dataset.pageTitle,
        locale: 'en'
    };

    window.remark_config = remarkConfig;

    (function (e, n) {
        for (var o = 0; o < e.length; o++) {
            var s = n.createElement('script'),
                c = '.js',
                d = n.head || n.body;
            if ('noModule' in s) {
                s.type = 'module';
                c = '.mjs';
            } else {
                s.async = true;
            }
            s.defer = true;
            s.src = window.remark_config.host + '/web/' + e[o] + c;
            d.appendChild(s);
        }
    })(remarkConfig.components || ['embed'], document);
})();
