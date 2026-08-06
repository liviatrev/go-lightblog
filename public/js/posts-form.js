(function () {
    var quill = new Quill('#editor-container', {
        theme: 'snow',
        placeholder: 'Start writing something amazing...',
        modules: {
            toolbar: [
                [{ header: [2, 3, 4, false] }],
                ['bold', 'italic', 'underline', 'strike'],
                ['blockquote', 'code-block'],
                [{ list: 'ordered' }, { list: 'bullet' }],
                ['link', 'image'],
                ['clean']
            ]
        }
    });

    var toolbarDOM = document.querySelector('.ql-toolbar');
    if (toolbarDOM) {
        var formatGroup = document.createElement('span');
        formatGroup.className = 'ql-formats';

        var customButton = document.createElement('button');
        customButton.innerHTML = '<b>\u003c/\u003e</b>';
        customButton.title = 'HTML / Code Mode';
        customButton.type = 'button';
        customButton.style.width = 'auto';
        customButton.style.padding = '0 5px';
        customButton.addEventListener('click', function (e) {
            e.preventDefault();
            toggleHtmlView();
        });

        formatGroup.appendChild(customButton);
        toolbarDOM.appendChild(formatGroup);
    }

    var isHtmlMode = false;
    var quillEditor = document.getElementById('editor-container');
    var htmlEditor = document.getElementById('html-editor');

    function toggleHtmlView() {
        if (isHtmlMode) {
            quill.root.innerHTML = htmlEditor.value;
            htmlEditor.style.display = 'none';
            quillEditor.style.display = 'block';
            isHtmlMode = false;
        } else {
            htmlEditor.value = quill.root.innerHTML;
            quillEditor.style.display = 'none';
            htmlEditor.style.display = 'block';
            isHtmlMode = true;
        }
    }

    var rawContentElement = document.getElementById('raw-content');
    if (rawContentElement) {
        quill.root.innerHTML = rawContentElement.textContent;
    }

    var form = document.getElementById('postForm');
    if (form) {
        form.onsubmit = function () {
            var contentInput = document.getElementById('hidden-content');
            if (isHtmlMode) {
                contentInput.value = htmlEditor.value;
            } else {
                contentInput.value = quill.root.innerHTML;
            }
        };
    }

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

    var generateSeoBtn = document.getElementById('generateSeoBtn');
    if (generateSeoBtn) {
        generateSeoBtn.addEventListener('click', function () {
            var content = quill.getText().trim();

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

    var typeSelect = document.getElementById('type');
    var divCategory = document.getElementById('div_category');
    var selectCategory = document.getElementById('select_category');
    var divTags = document.getElementById('div_tags');

    if (typeSelect && divCategory && selectCategory && divTags) {
        typeSelect.addEventListener('change', function () {
            if (this.value === 'page') {
                selectCategory.removeAttribute('required');
                divCategory.classList.add('d-none');
                divTags.classList.add('d-none');
            } else {
                selectCategory.setAttribute('required', '');
                divCategory.classList.remove('d-none');
                divTags.classList.remove('d-none');
            }
        });
    }
})();