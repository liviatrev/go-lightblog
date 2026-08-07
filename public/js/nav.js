/* go-lightblog - Lightweight navbar interactions (replaces bootstrap.bundle.min.js) */
(function () {
    'use strict';

    // ============================================
    // 1. Hamburger / Collapse Toggle
    // ============================================
    var toggler = document.querySelector('[data-bs-toggle="collapse"]');
    if (toggler) {
        var targetSelector = toggler.getAttribute('data-bs-target');
        var target = targetSelector ? document.querySelector(targetSelector) : null;

        toggler.addEventListener('click', function (e) {
            e.preventDefault();
            if (!target) return;

            var isExpanded = target.classList.contains('show');
            target.classList.toggle('show');
            toggler.classList.toggle('collapsed', isExpanded);
            toggler.setAttribute('aria-expanded', String(!isExpanded));
        });
    }

    // --- Close collapse when clicking a nav link (mobile UX) ---
    var navLinks = document.querySelectorAll('.navbar-collapse .nav-link, .navbar-collapse .dropdown-item');
    navLinks.forEach(function (link) {
        link.addEventListener('click', function () {
            var collapseEl = document.querySelector('.navbar-collapse.show');
            if (collapseEl) {
                collapseEl.classList.remove('show');
                var btn = document.querySelector('[data-bs-toggle="collapse"]');
                if (btn) {
                    btn.classList.add('collapsed');
                    btn.setAttribute('aria-expanded', 'false');
                }
            }
        });
    });

    // --- Close dropdown when clicking outside ---
    document.addEventListener('click', function (e) {
        var openDropdowns = document.querySelectorAll('.dropdown-menu.show');
        openDropdowns.forEach(function (menu) {
            var parent = menu.closest('.dropdown');
            if (parent && !parent.contains(e.target)) {
                menu.classList.remove('show');
                var toggle = parent.querySelector('.dropdown-toggle');
                if (toggle) toggle.setAttribute('aria-expanded', 'false');
            }
        });
    });

    // --- Dropdown Toggle ---
    var dropdownToggles = document.querySelectorAll('.dropdown-toggle');
    dropdownToggles.forEach(function (toggle) {
        toggle.addEventListener('click', function (e) {
            e.preventDefault();
            e.stopPropagation();
            var parent = toggle.closest('.dropdown');
            if (!parent) return;

            var menu = parent.querySelector('.dropdown-menu');
            if (!menu) return;

            var isOpen = menu.classList.contains('show');
            // Close all other open dropdowns
            document.querySelectorAll('.dropdown-menu.show').forEach(function (other) {
                if (other !== menu) {
                    other.classList.remove('show');
                    var otherToggle = other.closest('.dropdown').querySelector('.dropdown-toggle');
                    if (otherToggle) otherToggle.setAttribute('aria-expanded', 'false');
                }
            });

            menu.classList.toggle('show', !isOpen);
            toggle.setAttribute('aria-expanded', String(!isOpen));
        });
    });
})();