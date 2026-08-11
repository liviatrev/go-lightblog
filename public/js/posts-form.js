(function () {
    var rawContentElement = document.getElementById('raw-content');
    var editorContainer = document.getElementById('editor-container');

    if (rawContentElement && editorContainer) {
        editorContainer.innerHTML = rawContentElement.textContent; 
    }

    var editor = SUNEDITOR.create(document.getElementById('editor-container'), {
        plugins: SUNEDITOR.plugins,
        width: '100%',
        height: '420px',
        lang: SUNEDITOR_LANG.en,
        buttonList: [
            ['undo', 'redo'],
            ['font', 'fontSize', 'blockStyle'],
            ['bold', 'underline', 'italic', 'strike', 'subscript', 'superscript'],
            ['fontColor', 'textStyle', 'removeFormat'],
            ['outdent', 'indent', 'align', 'list', 'table'],
            ['link', 'image', 'video', 'lineHeight'],
            ['fullScreen', 'showBlocks', 'codeView'],
            ['preview', 'print'],
            ['-right', 'dir', 'paragraphStyle', 'blockquote'],
            ['-right', 'template', 'lineHeight'],
        ],
        placeholder: 'Start writing something amazing...'
    });

    var hiddenContent = document.getElementById('hidden-content');

    var form = document.getElementById('postForm');
    if (form) {
        form.onsubmit = function () {
            var content = '';
        
            var wysiwygDiv = document.querySelector('.sun-editor-editable');
            var codeTextarea = document.querySelector('.se-code-viewer');

            if (wysiwygDiv && codeTextarea) {
                if (window.getComputedStyle(wysiwygDiv).display === 'none') {
                    content = codeTextarea.value;
                } else {
                    content = wysiwygDiv.innerHTML;
                }
            }

            content = content.replace(/^<p><br><\/p>$/, '').trim();
            hiddenContent.value = content;
        };
    }

    var generateSeoBtn = document.getElementById('generateSeoBtn');
    if (generateSeoBtn) {
        generateSeoBtn.addEventListener('click', function () {
            var content = '';
        
            var wysiwygDiv = document.querySelector('.sun-editor-editable');
            var codeTextarea = document.querySelector('.se-code-viewer');

            if (wysiwygDiv && codeTextarea) {
                if (window.getComputedStyle(wysiwygDiv).display === 'none') {
                    content = codeTextarea.value;
                } else {
                    content = wysiwygDiv.innerHTML;
                }
            }

            content = content.replace(/^<p><br><\/p>$/, '');

            var tempDiv = document.createElement("div");
            tempDiv.innerHTML = content;

            content = tempDiv.textContent || tempDiv.innerText || "";

            content = content.trim();
            if (content.length < 50) {
                alert('Content is too short. Write at least a few sentences before pressing Generate AI.');
                return;
            }

            var btn = this;
            var originalBtnText = btn.innerHTML;

            btn.innerHTML = '⏳ Analyzing...';
            btn.classList.remove('btn-outline-success');
            btn.classList.add('btn-secondary');
            btn.disabled = true;

            fetch('/seo/generate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ content: content })
            })
                .then(function (response) {
                    if (!response.ok) throw new Error('Failed to contact AI');
                    return response.json();
                })
                .then(function (data) {
                    btn.innerHTML = originalBtnText;
                    btn.classList.remove('btn-secondary');
                    btn.classList.add('btn-outline-success');
                    btn.disabled = false;

                    if (data.error) {
                        alert('Error from AI: ' + data.error);
                    } else {
                        document.getElementById('metaTitle').value = data.meta_title || '';
                        document.getElementById('metaDesc').value = data.meta_description || '';
                        document.getElementById('targetKeyword').value = data.target_keyword || '';

                        document.getElementById('metaTitle').style.backgroundColor = '#e8f5e9';
                        setTimeout(function () {
                            document.getElementById('metaTitle').style.backgroundColor = '';
                        }, 1500);
                    }
                })
                .catch(function (error) {
                    console.error(error);
                    alert('Network error or AI limit reached.');
                    btn.innerHTML = originalBtnText;
                    btn.classList.remove('btn-secondary');
                    btn.classList.add('btn-outline-success');
                    btn.disabled = false;
                });
        });
    }

    // --- Cover image upload ---
    var uploadCoverBtn = document.getElementById('uploadCoverBtn');
    if (uploadCoverBtn) {
        uploadCoverBtn.addEventListener('change', function (e) {
            var file = e.target.files[0];
            if (!file) return;

            var statusText = document.getElementById('uploadStatus');
            statusText.innerText = 'Uploading to cloud...';
            statusText.className = 'form-text small text-info';

            var formData = new FormData();
            formData.append('image', file);

            fetch('/api/upload', {
                method: 'POST',
                body: formData
            })
                .then(function (response) { return response.json(); })
                .then(function (data) {
                    if (data.success) {
                        document.getElementById('coverImageURL').value = data.url;
                        statusText.innerText = 'Upload successful!';
                        statusText.className = 'form-text small text-success';
                    } else {
                        statusText.innerText = 'Failed: ' + data.message;
                        statusText.className = 'form-text small text-danger';
                    }
                })
                .catch(function (error) {
                    statusText.innerText = 'Network error occurred.';
                    statusText.className = 'form-text small text-danger';
                    console.error(error);
                });
        });
    }

    // --- Content Type selection toggling page/category/tags ---
    var typeSelect = document.getElementById('type');
    var divCategory = document.getElementById('div_category');
    var selectCategory = document.getElementById('select_category');
    var divTags = document.getElementById('div_tags');

    if (typeSelect) {
        typeSelect.addEventListener('change', function () {
            if (this.value === 'page') {
                if (selectCategory) selectCategory.removeAttribute('required');
                if (divCategory) divCategory.classList.add('d-none');
                if (divTags) divTags.classList.add('d-none');
            } else {
                if (selectCategory) selectCategory.setAttribute('required', '');
                if (divCategory) divCategory.classList.remove('d-none');
                if (divTags) divTags.classList.remove('d-none');
            }
        });
    }
})();