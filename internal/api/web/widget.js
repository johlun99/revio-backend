/* Revio — embeddable review widget
 * Usage:
 *   <script src="https://your-revio.example.com/widget.js"></script>
 *   <revio-reviews api-key="YOUR_KEY" product-id="YOUR_PRODUCT_ID"></revio-reviews>
 */
(function () {
  'use strict';

  // Capture script origin synchronously before any async code
  const _src = document.currentScript ? document.currentScript.src : '';
  const DEFAULT_BASE = _src ? new URL(_src).origin : window.location.origin;

  // ── Helpers ────────────────────────────────────────────────────────────────

  function esc(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function renderStars(rating, size) {
    var sz = size || 15;
    var html = '<span style="display:inline-flex;gap:2px;vertical-align:middle">';
    for (var i = 1; i <= 5; i++) {
      var filled = i <= Math.round(rating);
      html +=
        '<svg width="' + sz + '" height="' + sz + '" viewBox="0 0 24 24"' +
        ' fill="' + (filled ? 'var(--rv-star)' : 'none') + '"' +
        ' stroke="' + (filled ? 'var(--rv-star)' : 'var(--rv-star-empty)') + '"' +
        ' stroke-width="1.8" stroke-linejoin="round">' +
        '<polygon points="12,2 15.09,8.26 22,9.27 17,14.14 18.18,21.02 12,17.77 5.82,21.02 7,14.14 2,9.27 8.91,8.26"/>' +
        '</svg>';
    }
    return html + '</span>';
  }

  var MONTHS = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
  function formatDate(iso) {
    var d = new Date(iso);
    return MONTHS[d.getMonth()] + ' ' + d.getDate() + ', ' + d.getFullYear();
  }

  var AVATAR_COLORS = ['#6366f1','#8b5cf6','#ec4899','#f43f5e','#f59e0b','#10b981','#3b82f6','#0ea5e9'];
  function avatarColor(name) {
    var h = 0;
    for (var i = 0; i < name.length; i++) h = (h * 31 + name.charCodeAt(i)) | 0;
    return AVATAR_COLORS[Math.abs(h) % AVATAR_COLORS.length];
  }
  function initials(name) {
    return name.trim().split(/\s+/).map(function (w) { return w[0]; }).slice(0, 2).join('').toUpperCase();
  }

  // ── Styles ─────────────────────────────────────────────────────────────────

  var CSS = [
    ':host{display:block;font-family:-apple-system,BlinkMacSystemFont,"Inter","Segoe UI",sans-serif;font-size:15px;line-height:1.5;color:var(--rv-text,#111827);',
    '--rv-accent:#6366f1;--rv-accent-ring:rgba(99,102,241,.2);--rv-star:#f59e0b;--rv-star-empty:#d1d5db;',
    '--rv-text:#111827;--rv-text-muted:#6b7280;--rv-border:#e5e7eb;--rv-bg:#fff;--rv-bg-subtle:#f9fafb;--rv-radius:10px}',
    '*{box-sizing:border-box;margin:0;padding:0}',

    /* Summary */
    '.summary{display:flex;align-items:center;gap:20px;padding:20px 0 24px;border-bottom:1px solid var(--rv-border);margin-bottom:24px}',
    '.avg-score{font-size:52px;font-weight:700;letter-spacing:-3px;line-height:1;color:var(--rv-text)}',
    '.avg-meta{display:flex;flex-direction:column;gap:5px}',
    '.avg-count{font-size:13px;color:var(--rv-text-muted)}',

    /* Write button */
    '.write-btn{display:inline-flex;align-items:center;gap:7px;padding:9px 18px;background:var(--rv-accent);color:#fff;border:none;border-radius:var(--rv-radius);font-size:14px;font-weight:500;cursor:pointer;margin-bottom:24px;transition:opacity .15s;font-family:inherit}',
    '.write-btn:hover{opacity:.88}',

    /* Form */
    '.form-card{background:var(--rv-bg-subtle);border:1px solid var(--rv-border);border-radius:var(--rv-radius);padding:22px;margin-bottom:24px}',
    '.form-heading{font-size:15px;font-weight:600;margin-bottom:18px}',
    '.stars-picker{display:flex;gap:6px;margin-bottom:6px;cursor:pointer;user-select:none}',
    '.stars-picker .s{font-size:30px;line-height:1;color:var(--rv-star-empty);transition:color .1s,transform .1s}',
    '.stars-picker .s:hover,.stars-picker .s.on{color:var(--rv-star)}',
    '.stars-picker .s:hover{transform:scale(1.15)}',
    '.rating-hint{font-size:13px;color:var(--rv-text-muted);margin-bottom:14px}',
    '.field{margin-bottom:13px}',
    '.field label{display:block;font-size:13px;font-weight:500;color:var(--rv-text-muted);margin-bottom:4px}',
    '.field input,.field textarea{width:100%;padding:9px 12px;border:1px solid var(--rv-border);border-radius:7px;font-size:14px;font-family:inherit;background:var(--rv-bg);color:var(--rv-text);outline:none;transition:border-color .15s,box-shadow .15s}',
    '.field input:focus,.field textarea:focus{border-color:var(--rv-accent);box-shadow:0 0 0 3px var(--rv-accent-ring)}',
    '.field textarea{resize:vertical;min-height:88px}',
    '.form-row{display:grid;grid-template-columns:1fr 1fr;gap:12px}',
    '@media(max-width:480px){.form-row{grid-template-columns:1fr}}',
    '.form-actions{display:flex;gap:10px;margin-top:6px;align-items:center}',
    '.btn-primary{padding:9px 20px;background:var(--rv-accent);color:#fff;border:none;border-radius:7px;font-size:14px;font-weight:500;cursor:pointer;font-family:inherit;transition:opacity .15s}',
    '.btn-primary:hover{opacity:.88}',
    '.btn-primary:disabled{opacity:.45;cursor:not-allowed}',
    '.btn-ghost{padding:9px 14px;background:none;color:var(--rv-text-muted);border:1px solid var(--rv-border);border-radius:7px;font-size:14px;cursor:pointer;font-family:inherit;transition:background .15s}',
    '.btn-ghost:hover{background:var(--rv-border)}',
    '.form-error{font-size:13px;color:#dc2626;margin-top:10px}',

    /* Success banner */
    '.success-banner{background:#f0fdf4;border:1px solid #bbf7d0;border-radius:var(--rv-radius);padding:13px 16px;font-size:14px;color:#15803d;margin-bottom:24px}',

    /* Reviews */
    '.reviews-list{display:flex;flex-direction:column}',
    '.review-item{padding:20px 0;border-bottom:1px solid var(--rv-border)}',
    '.review-item:last-child{border-bottom:none}',
    '.review-header{display:flex;align-items:flex-start;gap:12px;margin-bottom:9px}',
    '.avatar{flex-shrink:0;width:38px;height:38px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:13px;font-weight:600;color:#fff}',
    '.review-meta{flex:1;min-width:0}',
    '.author-row{display:flex;align-items:center;gap:8px;flex-wrap:wrap;margin-bottom:3px}',
    '.author-name{font-weight:600;font-size:14px}',
    '.verified{display:inline-flex;align-items:center;gap:3px;font-size:11px;color:#059669;font-weight:500}',
    '.rating-date{display:flex;align-items:center;gap:8px}',
    '.review-date{font-size:12px;color:var(--rv-text-muted)}',
    '.review-title{font-weight:600;font-size:14px;margin-bottom:4px}',
    '.review-body{font-size:14px;color:#374151;line-height:1.65}',

    /* Load more */
    '.load-more-wrap{text-align:center;padding:22px 0 6px}',
    '.btn-load-more{padding:9px 26px;background:none;border:1px solid var(--rv-border);border-radius:7px;font-size:14px;color:var(--rv-text-muted);cursor:pointer;font-family:inherit;transition:background .15s,border-color .15s}',
    '.btn-load-more:hover{background:var(--rv-bg-subtle);border-color:#9ca3af}',
    '.btn-load-more:disabled{opacity:.5;cursor:not-allowed}',

    /* Empty / loading */
    '.empty{text-align:center;padding:36px 0;color:var(--rv-text-muted);font-size:14px}',
    '.loading-wrap{padding:48px 0;text-align:center}',
    '.spinner{display:inline-block;width:22px;height:22px;border:2px solid var(--rv-border);border-top-color:var(--rv-accent);border-radius:50%;animation:spin .7s linear infinite}',
    '@keyframes spin{to{transform:rotate(360deg)}}',
  ].join('');

  // ── Custom Element ──────────────────────────────────────────────────────────

  var RevioReviews = (function () {
    function RevioReviews() {
      var el = Reflect.construct(HTMLElement, [], RevioReviews);
      el._api = '';
      el._productId = '';
      el._base = DEFAULT_BASE;
      el._reviews = [];
      el._total = 0;
      el._avg = 0;
      el._offset = 0;
      el._limit = 10;
      el._loading = false;
      el._loadingMore = false;
      el._formVisible = false;
      el._formRating = 0;
      el._formHover = 0;
      el._submitState = 'idle'; // idle | submitting | success | error
      el._submitError = '';
      return el;
    }
    RevioReviews.prototype = Object.create(HTMLElement.prototype, { constructor: { value: RevioReviews } });
    Object.setPrototypeOf(RevioReviews, HTMLElement);

    RevioReviews.prototype.connectedCallback = function () {
      this._api = this.getAttribute('api-key') || '';
      this._productId = this.getAttribute('product-id') || '';
      this._base = this.getAttribute('api-url') || DEFAULT_BASE;

      var root = this.attachShadow({ mode: 'open' });
      var self = this;
      root.addEventListener('click', function (e) { self._onClick(e); });
      root.addEventListener('submit', function (e) { self._onSubmit(e); });
      root.addEventListener('mouseover', function (e) { self._onStarHover(e); });
      root.addEventListener('mouseout', function (e) { self._onStarOut(e); });

      this._load(true);
    };

    RevioReviews.prototype._load = function (reset) {
      if (this._loading || this._loadingMore) return;
      var self = this;
      if (reset) {
        this._reviews = [];
        this._offset = 0;
        this._loading = true;
      } else {
        this._loadingMore = true;
      }
      this._update();

      var qs = 'product_id=' + encodeURIComponent(this._productId) +
               '&limit=' + this._limit +
               '&offset=' + this._offset;

      fetch(this._base + '/api/v1/reviews?' + qs, { headers: { 'X-API-Key': this._api } })
        .then(function (res) {
          if (!res.ok) throw new Error('HTTP ' + res.status);
          return res.json();
        })
        .then(function (json) {
          var batch = json.data || [];
          self._reviews = reset ? batch : self._reviews.concat(batch);
          self._total = json.total || 0;
          self._avg = json.avg_rating || 0;
          self._offset = self._reviews.length;
        })
        .catch(function (e) { console.error('[revio-reviews]', e); })
        .finally(function () {
          self._loading = false;
          self._loadingMore = false;
          self._update();
        });
    };

    RevioReviews.prototype._submit = function (form) {
      if (!this._formRating) return;
      var self = this;
      var fd = new FormData(form);

      this._submitState = 'submitting';
      this._update();

      var body = {
        product_id: this._productId,
        author_name: fd.get('author_name'),
        rating: this._formRating,
        body: fd.get('body'),
      };
      var email = (fd.get('author_email') || '').trim();
      if (email) body.author_email = email;
      var title = (fd.get('title') || '').trim();
      if (title) body.title = title;

      fetch(this._base + '/api/v1/reviews', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-API-Key': this._api },
        body: JSON.stringify(body),
      })
        .then(function (res) {
          if (!res.ok) return res.json().then(function (e) { throw new Error(e.error || 'HTTP ' + res.status); });
          self._submitState = 'success';
          self._formVisible = false;
          self._formRating = 0;
          self._load(true);
        })
        .catch(function (e) {
          self._submitState = 'error';
          self._submitError = e.message;
          self._update();
        });
    };

    RevioReviews.prototype._onClick = function (e) {
      if (e.target.closest('[data-action="show-form"]')) {
        this._formVisible = true;
        this._submitState = 'idle';
        this._update();
        var self = this;
        requestAnimationFrame(function () {
          var el = self.shadowRoot.querySelector('input[name=author_name]');
          if (el) el.focus();
        });
        return;
      }
      if (e.target.closest('[data-action="hide-form"]')) {
        this._formVisible = false;
        this._update();
        return;
      }
      if (e.target.closest('[data-action="load-more"]')) {
        this._load(false);
        return;
      }
      var star = e.target.closest('[data-star]');
      if (star) {
        this._formRating = parseInt(star.dataset.star, 10);
        this._updateStarPicker();
      }
    };

    RevioReviews.prototype._onSubmit = function (e) {
      e.preventDefault();
      this._submit(e.target);
    };

    RevioReviews.prototype._onStarHover = function (e) {
      var star = e.target.closest('[data-star]');
      if (star && star.closest('[data-picker]')) {
        this._formHover = parseInt(star.dataset.star, 10);
        this._updateStarPicker();
      }
    };

    RevioReviews.prototype._onStarOut = function (e) {
      if (e.target.closest('[data-picker]')) {
        this._formHover = 0;
        this._updateStarPicker();
      }
    };

    RevioReviews.prototype._updateStarPicker = function () {
      var picker = this.shadowRoot.querySelector('[data-picker]');
      if (!picker) return;
      var active = this._formHover || this._formRating;
      picker.querySelectorAll('[data-star]').forEach(function (el) {
        var n = parseInt(el.dataset.star, 10);
        el.classList.toggle('on', n <= active);
        el.textContent = n <= active ? '★' : '☆';
      });
    };

    RevioReviews.prototype._update = function () {
      // Preserve form text state across re-renders
      var prev = this.shadowRoot.querySelector('form');
      var saved = prev ? {
        author_name: (prev.querySelector('[name=author_name]') || {}).value || '',
        author_email: (prev.querySelector('[name=author_email]') || {}).value || '',
        title: (prev.querySelector('[name=title]') || {}).value || '',
        body: (prev.querySelector('[name=body]') || {}).value || '',
      } : null;

      this.shadowRoot.innerHTML = '<style>' + CSS + '</style>' + this._html();

      if (saved) {
        var form = this.shadowRoot.querySelector('form');
        if (form) {
          Object.keys(saved).forEach(function (k) {
            var el = form.querySelector('[name=' + k + ']');
            if (el) el.value = saved[k];
          });
        }
      }
    };

    RevioReviews.prototype._html = function () {
      if (this._loading) {
        return '<div class="loading-wrap"><div class="spinner"></div></div>';
      }

      var parts = [];

      if (this._total > 0 || this._reviews.length > 0) {
        parts.push(
          '<div class="summary">' +
          '<div class="avg-score">' + this._avg.toFixed(1) + '</div>' +
          '<div class="avg-meta">' +
          renderStars(this._avg, 18) +
          '<div class="avg-count">' + this._total + ' review' + (this._total !== 1 ? 's' : '') + '</div>' +
          '</div></div>'
        );
      }

      if (this._formVisible) {
        parts.push(this._formHTML());
      } else {
        parts.push(
          '<button class="write-btn" data-action="show-form">' +
          '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">' +
          '<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>' +
          '<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>' +
          '</svg>Write a review</button>'
        );
      }

      if (this._submitState === 'success') {
        parts.push('<div class="success-banner">✓ Thank you! Your review has been submitted and is pending approval.</div>');
      }

      if (this._reviews.length === 0) {
        parts.push('<div class="empty">No reviews yet — be the first to share your experience.</div>');
      } else {
        var rows = this._reviews.map(this._reviewHTML.bind(this)).join('');
        parts.push('<div class="reviews-list">' + rows + '</div>');
      }

      if (this._reviews.length < this._total) {
        var remaining = this._total - this._reviews.length;
        parts.push(
          '<div class="load-more-wrap">' +
          '<button class="btn-load-more" data-action="load-more"' + (this._loadingMore ? ' disabled' : '') + '>' +
          (this._loadingMore
            ? '<div class="spinner" style="width:16px;height:16px;border-width:2px;display:inline-block"></div>'
            : 'Show more (' + remaining + ' remaining)') +
          '</button></div>'
        );
      }

      return parts.join('');
    };

    RevioReviews.prototype._formHTML = function () {
      var active = this._formHover || this._formRating;
      var starsHTML = [1,2,3,4,5].map(function (n) {
        return '<span class="s' + (n <= active ? ' on' : '') + '" data-star="' + n + '">' + (n <= active ? '★' : '☆') + '</span>';
      }).join('');

      var submitDisabled = this._submitState === 'submitting' || !this._formRating;

      return (
        '<div class="form-card">' +
        '<div class="form-heading">Write a review</div>' +
        '<div class="stars-picker" data-picker>' + starsHTML + '</div>' +
        (!this._formRating ? '<div class="rating-hint">Select a star rating to continue</div>' : '') +
        '<form>' +
        '<div class="form-row">' +
        '<div class="field"><label>Name <span style="color:#ef4444">*</span></label><input name="author_name" type="text" placeholder="Your name" required></div>' +
        '<div class="field"><label>Email <span style="color:var(--rv-text-muted);font-weight:400">(optional)</span></label><input name="author_email" type="email" placeholder="your@email.com"></div>' +
        '</div>' +
        '<div class="field"><label>Title <span style="color:var(--rv-text-muted);font-weight:400">(optional)</span></label><input name="title" type="text" placeholder="Summarise your experience"></div>' +
        '<div class="field"><label>Review <span style="color:#ef4444">*</span></label><textarea name="body" placeholder="What did you think of the product?" required></textarea></div>' +
        '<div class="form-actions">' +
        '<button type="submit" class="btn-primary"' + (submitDisabled ? ' disabled' : '') + '>' +
        (this._submitState === 'submitting' ? 'Submitting…' : 'Submit review') +
        '</button>' +
        '<button type="button" class="btn-ghost" data-action="hide-form">Cancel</button>' +
        '</div>' +
        (this._submitState === 'error' ? '<div class="form-error">' + esc(this._submitError) + '</div>' : '') +
        '</form></div>'
      );
    };

    RevioReviews.prototype._reviewHTML = function (r) {
      var color = avatarColor(r.author_name);
      var ini = initials(r.author_name);
      var date = formatDate(r.created_at);

      return (
        '<div class="review-item">' +
        '<div class="review-header">' +
        '<div class="avatar" style="background:' + color + '">' + esc(ini) + '</div>' +
        '<div class="review-meta">' +
        '<div class="author-row">' +
        '<span class="author-name">' + esc(r.author_name) + '</span>' +
        (r.verified_purchase
          ? '<span class="verified"><svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/></svg>Verified purchase</span>'
          : '') +
        '</div>' +
        '<div class="rating-date">' + renderStars(r.rating, 13) + '<span class="review-date">' + date + '</span></div>' +
        '</div></div>' +
        (r.title ? '<div class="review-title">' + esc(r.title) + '</div>' : '') +
        '<div class="review-body">' + esc(r.body) + '</div>' +
        '</div>'
      );
    };

    return RevioReviews;
  }());

  if (!customElements.get('revio-reviews')) {
    customElements.define('revio-reviews', RevioReviews);
  }
}());
