    const tbody = document.getElementById('pozycje-tbody');
    const btnDodaj = document.getElementById('btn-dodaj');
    const btnGeneruj = document.getElementById('btn-generuj');
    const form = document.getElementById('oferta-form');
    const message = document.getElementById('message');
    var accordionFormReady = false;

    const shareSupported = typeof navigator !== 'undefined'
      && typeof navigator.canShare === 'function'
      && typeof navigator.share === 'function';

    const STORAGE_KEY_SOURCE = 'sumit_source';
    const PAGE_SIGNATURE = 'Franciszek Dranka';

    function createPageSignatureElement(extraClass) {
      const p = document.createElement('p');
      p.className = extraClass ? 'page-signature ' + extraClass : 'page-signature';
      p.setAttribute('aria-hidden', 'true');
      p.textContent = PAGE_SIGNATURE;
      return p;
    }

    function zachowajPozycjeScroll(fn) {
      const y = window.scrollY || document.documentElement.scrollTop || 0;
      fn();
      requestAnimationFrame(() => {
        window.scrollTo(0, y);
      });
    }

    (function initTrackingSource() {
      try {
        const src = new URLSearchParams(location.search).get('src');
        if (src) localStorage.setItem(STORAGE_KEY_SOURCE, src);
      } catch (_) {}
    })();

    function getTrackingSource() {
      try {
        return localStorage.getItem(STORAGE_KEY_SOURCE) || '';
      } catch (_) {
        return '';
      }
    }

    function trackEvent(eventName) {
      fetch('/api/track', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ event: eventName, source: getTrackingSource() }),
      }).catch(function () {});
    }

    const STORAGE_KEY_THEME = 'sumit_theme';
    const STORAGE_KEY_APP_PREFS = 'sumit_app_prefs';
    const DEFAULT_VALIDITY_DAYS = 14;
    const VALIDITY_DAY_OPTIONS = [7, 14, 30];
    const DEFAULT_DOC_TYPE_OPTIONS = ['', 'faktura_vat'];
    const DEFAULT_VAT_OPTIONS = ['23', '8', '5', '0'];
    const btnTheme = document.getElementById('btn-theme');
    const btnThemeMobile = document.getElementById('btn-theme-mobile');
    const themeButtons = [btnTheme, btnThemeMobile].filter(Boolean);

    function wczytajPreferencjeMotywu() {
      try {
        const zapis = localStorage.getItem(STORAGE_KEY_THEME);
        if (zapis === 'dark' || zapis === 'light') return zapis;
      } catch (e) {}
      return 'system';
    }

    function wczytajDomyslnaWaznoscDni() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY_APP_PREFS);
        if (!raw) return DEFAULT_VALIDITY_DAYS;
        const dane = JSON.parse(raw);
        const n = Number(dane && dane.defaultValidityDays);
        if (VALIDITY_DAY_OPTIONS.includes(n)) return n;
      } catch (e) {}
      return DEFAULT_VALIDITY_DAYS;
    }

    function zapiszDomyslnaWaznoscDni(dni) {
      if (!VALIDITY_DAY_OPTIONS.includes(dni)) return;
      try {
        const raw = localStorage.getItem(STORAGE_KEY_APP_PREFS);
        const dane = raw ? JSON.parse(raw) : {};
        dane.defaultValidityDays = dni;
        localStorage.setItem(STORAGE_KEY_APP_PREFS, JSON.stringify(dane));
      } catch (e) {}
    }

    function wczytajDomyslnyTypDokumentu() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY_APP_PREFS);
        if (!raw) return '';
        const dane = JSON.parse(raw);
        const typ = typeof dane.defaultDocType === 'string' ? dane.defaultDocType : '';
        return DEFAULT_DOC_TYPE_OPTIONS.includes(typ) ? typ : '';
      } catch (e) {}
      return '';
    }

    function zapiszDomyslnyTypDokumentu(typ) {
      if (!DEFAULT_DOC_TYPE_OPTIONS.includes(typ)) return;
      try {
        const raw = localStorage.getItem(STORAGE_KEY_APP_PREFS);
        const dane = raw ? JSON.parse(raw) : {};
        dane.defaultDocType = typ;
        localStorage.setItem(STORAGE_KEY_APP_PREFS, JSON.stringify(dane));
      } catch (e) {}
    }

    function wczytajDomyslnaStawkeVat() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY_APP_PREFS);
        if (!raw) return '23';
        const dane = JSON.parse(raw);
        const vat = String(dane.defaultVatRate || '23');
        return DEFAULT_VAT_OPTIONS.includes(vat) ? vat : '23';
      } catch (e) {}
      return '23';
    }

    function zapiszDomyslnaStawkeVat(vat) {
      if (!DEFAULT_VAT_OPTIONS.includes(vat)) return;
      try {
        const raw = localStorage.getItem(STORAGE_KEY_APP_PREFS);
        const dane = raw ? JSON.parse(raw) : {};
        dane.defaultVatRate = vat;
        localStorage.setItem(STORAGE_KEY_APP_PREFS, JSON.stringify(dane));
      } catch (e) {}
    }

    function ustawPreferencjeMotywu(pref) {
      if (pref === 'system') {
        try { localStorage.removeItem(STORAGE_KEY_THEME); } catch (e) {}
        const dark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
        ustawMotyw(dark ? 'dark' : 'light', false);
        return;
      }
      ustawMotyw(pref === 'dark' ? 'dark' : 'light', true);
    }

    function odswiezChipyMotywuApp() {
      const wrap = document.getElementById('app-settings-theme');
      if (!wrap) return;
      const pref = wczytajPreferencjeMotywu();
      wrap.querySelectorAll('.chip[data-theme-pref]').forEach((chip) => {
        const active = chip.getAttribute('data-theme-pref') === pref;
        chip.classList.toggle('is-active', active);
        chip.setAttribute('aria-pressed', active ? 'true' : 'false');
      });
    }

    function odswiezChipyWaznosciApp() {
      const wrap = document.getElementById('app-settings-validity');
      if (!wrap) return;
      const dni = wczytajDomyslnaWaznoscDni();
      wrap.querySelectorAll('.chip[data-validity-days]').forEach((chip) => {
        const active = Number(chip.getAttribute('data-validity-days')) === dni;
        chip.classList.toggle('is-active', active);
        chip.setAttribute('aria-pressed', active ? 'true' : 'false');
      });
    }

    function odswiezChipyDocTypeApp() {
      const wrap = document.getElementById('app-settings-doc-type');
      if (!wrap) return;
      const typ = wczytajDomyslnyTypDokumentu();
      wrap.querySelectorAll('.chip[data-default-doc-type]').forEach((chip) => {
        const active = (chip.getAttribute('data-default-doc-type') || '') === typ;
        chip.classList.toggle('is-active', active);
        chip.setAttribute('aria-pressed', active ? 'true' : 'false');
      });
    }

    function odswiezChipyVatApp() {
      const wrap = document.getElementById('app-settings-vat');
      if (!wrap) return;
      const vat = wczytajDomyslnaStawkeVat();
      wrap.querySelectorAll('.chip[data-default-vat]').forEach((chip) => {
        const active = chip.getAttribute('data-default-vat') === vat;
        chip.classList.toggle('is-active', active);
        chip.setAttribute('aria-pressed', active ? 'true' : 'false');
      });
    }

    function aktualnyMotyw() {
      return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light';
    }

    function aktualizujOpisPrzeciskaMotywu(motyw) {
      const opis = motyw === 'dark' ? 'Włącz tryb jasny' : 'Włącz tryb ciemny';
      themeButtons.forEach((btn) => {
        btn.setAttribute('aria-label', opis);
        btn.setAttribute('title', opis);
      });
    }

    function ustawMotyw(motyw, zapisz) {
      const nowy = motyw === 'dark' ? 'dark' : 'light';
      document.documentElement.setAttribute('data-theme', nowy);
      aktualizujOpisPrzeciskaMotywu(nowy);
      if (zapisz) {
        try { localStorage.setItem(STORAGE_KEY_THEME, nowy); } catch (e) {}
      }
    }

    aktualizujOpisPrzeciskaMotywu(aktualnyMotyw());

    themeButtons.forEach((btn) => {
      btn.addEventListener('click', () => {
        const nowy = aktualnyMotyw() === 'dark' ? 'light' : 'dark';
        ustawMotyw(nowy, true);
        odswiezChipyMotywuApp();
      });
    });

    if (window.matchMedia) {
      const mql = window.matchMedia('(prefers-color-scheme: dark)');
      const handler = (e) => {
        let zapis = null;
        try { zapis = localStorage.getItem(STORAGE_KEY_THEME); } catch (err) {}
        if (zapis !== 'dark' && zapis !== 'light') {
          ustawMotyw(e.matches ? 'dark' : 'light', false);
          odswiezChipyMotywuApp();
        }
      };
      if (mql.addEventListener) mql.addEventListener('change', handler);
      else if (mql.addListener) mql.addListener(handler);
    }

    function odswiezNumery() {
      [...tbody.querySelectorAll('tr')].forEach((tr, i) => {
        tr.querySelector('.col-lp').textContent = i + 1;
      });
    }

    function dodajWiersz(opts) {
      const naKoniec = !!(opts && opts.naKoniec);
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td class="col-lp"></td>
        <td class="col-nazwa"><input type="text" class="in-nazwa" placeholder="np. Usługa konsultingowa" list="szablony-pozycji-datalist" autocomplete="off" required></td>
        <td class="col-ilosc">
          <div class="ilosc-wrap">
            <input type="number" class="in-ilosc" step="any" min="0" placeholder="1" required>
            <button type="button" class="btn-calc" title="Kalkulator powierzchni (m²)" aria-label="Otwórz kalkulator powierzchni" aria-haspopup="dialog" aria-controls="kalkulator-modal">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <rect x="4" y="2" width="16" height="20" rx="2"></rect>
                <line x1="8" y1="6" x2="16" y2="6"></line>
                <line x1="8" y1="10" x2="8" y2="10"></line>
                <line x1="12" y1="10" x2="12" y2="10"></line>
                <line x1="16" y1="10" x2="16" y2="10"></line>
                <line x1="8" y1="14" x2="8" y2="14"></line>
                <line x1="12" y1="14" x2="12" y2="14"></line>
                <line x1="16" y1="14" x2="16" y2="18"></line>
                <line x1="8" y1="18" x2="12" y2="18"></line>
              </svg>
            </button>
          </div>
        </td>
        <td class="col-cena">
          <div class="price-wrap">
            <input type="number" class="in-cena" step="0.01" min="0" placeholder="0,00" aria-label="Cena dla klienta (zł)" required>
            <div class="koszt-row">
              <span class="koszt-label" aria-hidden="true">Twój koszt (za szt.)</span>
              <input type="number" class="in-koszt" step="0.01" min="0" placeholder="0,00" inputmode="decimal" aria-label="Twój koszt za jednostkę" title="Ile Ciebie kosztuje 1 jednostka — służy tylko do liczenia szacowanego zysku, nie trafia na PDF">
            </div>
            <div class="price-adjusts" role="group" aria-label="Szybka korekta ceny jednostkowej">
              <button type="button" class="btn-adjust" data-adjust="-5" title="Obniż cenę jednostkową o 5%" aria-label="Obniż cenę jednostkową o 5% (rabat)">Rab. -5%</button>
              <button type="button" class="btn-adjust" data-adjust="10" title="Dolicz 10% marży do ceny jednostkowej" aria-label="Dolicz 10% marży do ceny jednostkowej">Mar. +10%</button>
            </div>
          </div>
        </td>
        <td class="col-vat vat-col-hidden">
          <select class="in-vat" aria-label="Stawka VAT">
            <option value="23">23%</option>
            <option value="8">8%</option>
            <option value="5">5%</option>
            <option value="0">0%</option>
          </select>
        </td>
        <td class="col-akcja">
          <div class="row-actions">
            <button type="button" class="btn-duplicate" title="Duplikuj wiersz" aria-label="Duplikuj wiersz">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
              </svg>
            </button>
            <button type="button" class="btn-remove" title="Usuń" aria-label="Usuń wiersz">&times;</button>
          </div>
        </td>
      `;
      tr.querySelector('.btn-remove').addEventListener('click', () => {
        if (tbody.children.length <= 1) {
          pokazKomunikat('Musi pozostać co najmniej jedna pozycja.', 'error');
          return;
        }
        tr.remove();
        odswiezNumery();
        aktualizujSzacowanyZysk();
        odswiezPodgladPozycjiAkordeon();
      });
      tr.querySelector('.btn-duplicate').addEventListener('click', () => duplikujWiersz(tr));
      tr.querySelector('.btn-calc').addEventListener('click', () => otworzKalkulator(tr));
      tr.querySelectorAll('.btn-adjust').forEach(btn => {
        btn.addEventListener('click', () => {
          const pct = parseFloat(btn.dataset.adjust);
          if (Number.isFinite(pct)) applyAdjustment(tr, pct);
        });
      });
      const nazwaInput = tr.querySelector('.in-nazwa');
      if (nazwaInput) {
        nazwaInput.addEventListener('input', () => aplikujSzablonDoWiersza(tr));
      }
      if (naKoniec) {
        tbody.appendChild(tr);
      } else {
        tbody.insertBefore(tr, tbody.firstChild);
      }
      odswiezNumery();
      if (accordionFormReady) odswiezPodgladPozycjiAkordeon();
      const inVat = tr.querySelector('.in-vat');
      if (inVat) inVat.value = wczytajDomyslnaStawkeVat();
      const inNazwa = tr.querySelector('.in-nazwa');
      if (!naKoniec && inNazwa && document.activeElement === btnDodaj) {
        inNazwa.focus();
      }
    }

    function duplikujWiersz(originalTr) {
      if (!originalTr || !tbody.contains(originalTr)) return;
      const nazwa = originalTr.querySelector('.in-nazwa').value;
      const ilosc = originalTr.querySelector('.in-ilosc').value;
      const cena  = originalTr.querySelector('.in-cena').value;
      const koszt = originalTr.querySelector('.in-koszt').value;
      const vat   = originalTr.querySelector('.in-vat') ? originalTr.querySelector('.in-vat').value : '23';

      dodajWiersz();
      const noweTr = tbody.firstElementChild;
      if (!noweTr) return;
      if (noweTr !== originalTr) {
        originalTr.insertAdjacentElement('afterend', noweTr);
      }

      noweTr.querySelector('.in-nazwa').value = nazwa;
      noweTr.querySelector('.in-ilosc').value = ilosc;
      noweTr.querySelector('.in-cena').value  = cena;
      noweTr.querySelector('.in-koszt').value = koszt;
      if (noweTr.querySelector('.in-vat')) noweTr.querySelector('.in-vat').value = vat;

      odswiezNumery();
      aktualizujSzacowanyZysk();
      saveDraft();
    }

    function applyAdjustment(row, percentage) {
      if (!row) return;
      const input = row.querySelector('.in-cena');
      if (!input) return;
      const raw = String(input.value).replace(',', '.').trim();
      const current = parseFloat(raw);
      if (!Number.isFinite(current)) {
        pokazKomunikat('Podaj cenę jednostkową, zanim zastosujesz korektę.', 'error');
        input.focus();
        return;
      }
      const next = current * (1 + percentage / 100);
      const rounded = Math.max(0, Math.round(next * 100) / 100);
      input.value = String(rounded);
      input.dispatchEvent(new Event('input', { bubbles: true }));
    }

    function pokazKomunikat(tekst, typ) {
      message.textContent = tekst;
      message.className = 'message ' + typ;
    }

    function ukryjKomunikat() {
      message.textContent = '';
      message.className = 'message';
    }

    const profitBadge = document.getElementById('profit-badge');
    const profitBadgeValue = document.getElementById('profit-badge-value');

    function parsujLiczbeInput(el) {
      if (!el) return NaN;
      const raw = String(el.value || '').replace(',', '.').trim();
      if (!raw) return NaN;
      return parseFloat(raw);
    }

    function formatujKwote(v) {
      const n = Number(v);
      if (!Number.isFinite(n)) return '0,00 zł';
      return n.toFixed(2).replace('.', ',') + ' zł';
    }

    function aktualizujSzacowanyZysk() {
      if (!profitBadge || !profitBadgeValue) return;
      let suma = 0;
      let maPelneDane = false;
      [...tbody.querySelectorAll('tr')].forEach(tr => {
        const ilosc = parsujLiczbeInput(tr.querySelector('.in-ilosc'));
        const cena  = parsujLiczbeInput(tr.querySelector('.in-cena'));
        const koszt = parsujLiczbeInput(tr.querySelector('.in-koszt'));
        if (!Number.isFinite(ilosc) || !Number.isFinite(cena) || !Number.isFinite(koszt)) return;
        suma += (cena - koszt) * ilosc;
        maPelneDane = true;
      });
      profitBadgeValue.textContent = formatujKwote(suma);
      profitBadge.classList.remove('is-zero', 'is-loss');
      if (!maPelneDane || Math.abs(suma) < 0.005) {
        profitBadge.classList.add('is-zero');
      } else if (suma < 0) {
        profitBadge.classList.add('is-loss');
      }
    }

    tbody.addEventListener('input', (e) => {
      if (!e.target) return;
      if (e.target.classList && (
        e.target.classList.contains('in-ilosc') ||
        e.target.classList.contains('in-cena') ||
        e.target.classList.contains('in-koszt')
      )) {
        aktualizujSzacowanyZysk();
      }
      if (e.target.classList && e.target.classList.contains('in-nazwa')) {
        odswiezPodgladPozycjiAkordeon();
      }
    });

    btnDodaj.addEventListener('click', dodajWiersz);

    dodajWiersz();
    aktualizujSzacowanyZysk();

    (function ustawDomyslnaDateWaznosciPrzyStarcie() {
      const input = document.getElementById('data_waznosci');
      if (!input || input.value) return;
      const d = new Date();
      d.setDate(d.getDate() + wczytajDomyslnaWaznoscDni());
      const yyyy = d.getFullYear();
      const mm = String(d.getMonth() + 1).padStart(2, '0');
      const dd = String(d.getDate()).padStart(2, '0');
      input.value = `${yyyy}-${mm}-${dd}`;
    })();

    function ustawDateZaDniOdDzis(dni) {
      const input = document.getElementById('data_waznosci');
      if (!input) return;
      const d = new Date();
      d.setDate(d.getDate() + dni);
      const yyyy = d.getFullYear();
      const mm = String(d.getMonth() + 1).padStart(2, '0');
      const dd = String(d.getDate()).padStart(2, '0');
      input.value = `${yyyy}-${mm}-${dd}`;
      input.dispatchEvent(new Event('input', { bubbles: true }));
    }

    document.querySelectorAll('.date-helpers .chip').forEach(btn => {
      btn.addEventListener('click', () => {
        const dni = parseInt(btn.dataset.days, 10);
        if (Number.isFinite(dni)) ustawDateZaDniOdDzis(dni);
      });
    });

    const uwagiTextarea = document.getElementById('uwagi');
    const presetChips = document.querySelectorAll('.footer-presets .chip');

    function presetLines() {
      if (!uwagiTextarea) return [];
      return uwagiTextarea.value.split('\n');
    }

    function togglePreset(text) {
      if (!uwagiTextarea) return;
      const target = String(text).trim();
      if (!target) return;

      zachowajPozycjeScroll(() => {
        const lines = presetLines();
        const idx = lines.findIndex(l => l.trim() === target);

        if (idx !== -1) {
          lines.splice(idx, 1);
          while (lines.length && lines[lines.length - 1].trim() === '') lines.pop();
          uwagiTextarea.value = lines.join('\n');
        } else {
          const current = uwagiTextarea.value;
          const sep = current && !current.endsWith('\n') ? '\n' : '';
          uwagiTextarea.value = current + sep + target;
        }

        uwagiTextarea.dispatchEvent(new Event('input', { bubbles: true }));
        odswiezStanPresetow();
      });
    }

    function odswiezStanPresetow() {
      if (!uwagiTextarea) return;
      const lines = presetLines().map(l => l.trim());
      presetChips.forEach(chip => {
        const preset = (chip.dataset.preset || '').trim();
        const aktywny = preset && lines.includes(preset);
        chip.setAttribute('aria-pressed', aktywny ? 'true' : 'false');
      });
    }

    presetChips.forEach(chip => {
      chip.addEventListener('click', () => togglePreset(chip.dataset.preset || ''));
    });

    if (uwagiTextarea) {
      uwagiTextarea.addEventListener('input', odswiezStanPresetow);
    }

    function budujPayloadZFormularza() {
      const cfg = wczytajConfig();
      const nazwaFirmy = String(cfg.nazwa_firmy || '').trim();
      const klientEl = document.getElementById('klient');
      const klient = klientEl ? klientEl.value.trim() : '';
      const pozycjeWiersze = [...tbody.querySelectorAll('tr')];
      // Uwaga: pole .in-koszt jest celowo pomijane w payloadzie /quote.
      // Backend Go (struct Pozycja w oferta.go) nie zna pola "koszt" — dorzucenie go
      // zepsułoby parsowanie JSON-a po stronie serwera. Koszt własny pozostaje tylko
      // w warstwie UI i służy wyłącznie do liczenia szacowanego zysku w przeglądarce
      // oraz jest zapisywany razem z wpisem historii (osobno od payloadu) na potrzeby
      // statystyk zysku.
      const aktualnyDocType = (document.getElementById('doc-type-switcher') &&
        document.querySelector('#doc-type-switcher .chip.is-active'))
        ? (document.querySelector('#doc-type-switcher .chip.is-active').dataset.docType || '')
        : '';

      const pozycje = pozycjeWiersze.map(tr => {
        const p = {
          nazwa: tr.querySelector('.in-nazwa').value.trim(),
          ilosc: parseFloat(tr.querySelector('.in-ilosc').value),
          cena_jednostkowa: parseFloat(tr.querySelector('.in-cena').value),
        };
        if (aktualnyDocType === 'faktura_vat') {
          const vatEl = tr.querySelector('.in-vat');
          p.stawka_vat = vatEl ? parseFloat(vatEl.value) || 0 : 23;
        }
        return p;
      });
      const koszty = pozycjeWiersze.map(tr => {
        const raw = String(tr.querySelector('.in-koszt').value || '').replace(',', '.').trim();
        if (!raw) return null;
        const v = parseFloat(raw);
        return Number.isFinite(v) ? v : null;
      });
      const pozycjeOk = pozycje.filter(p => p.nazwa && Number.isFinite(p.ilosc) && p.ilosc > 0 && Number.isFinite(p.cena_jednostkowa) && p.cena_jednostkowa >= 0);

      const numerFaktury = document.getElementById('numer_faktury') ? document.getElementById('numer_faktury').value.trim() : '';
      const dataSprzedazy = document.getElementById('data_sprzedazy') ? document.getElementById('data_sprzedazy').value : '';
      const terminPlatnosci = document.getElementById('termin_platnosci') ? document.getElementById('termin_platnosci').value.trim() : '';
      const isInvoice = aktualnyDocType === 'faktura_proforma' || aktualnyDocType === 'faktura_vat';
      const invoiceReady = !!numerFaktury && !!dataSprzedazy;

      const gotowy = !!nazwaFirmy && !!klient && pozycjeOk.length > 0;

      const payloadBase = {
        nazwa_firmy: nazwaFirmy,
        nip: String(cfg.nip || '').trim(),
        adres: String(cfg.adres || '').trim(),
        miasto: String(cfg.miasto || '').trim(),
        telefon: String(cfg.telefon || '').trim(),
        email: String(cfg.email || '').trim(),
        logo_base64: String(cfg.logo_base64 || ''),
        numer_konta: String(cfg.numer_konta || '').trim(),
        klient: klient,
        numer_oferty: document.getElementById('numer_oferty').value.trim(),
        data_waznosci: document.getElementById('data_waznosci').value,
        uwagi: document.getElementById('uwagi').value.trim(),
        pozycje: pozycjeOk,
      };
      if (aktualnyDocType && (!isInvoice || invoiceReady)) {
        payloadBase.typ_dokumentu = aktualnyDocType;
        payloadBase.numer_faktury = numerFaktury;
        payloadBase.data_sprzedazy = dataSprzedazy;
        payloadBase.termin_platnosci = terminPlatnosci;
      }

      return {
        gotowy,
        nazwaFirmy,
        klient,
        koszty,
        payload: payloadBase,
      };
    }

    const livePodgladMql = (window.matchMedia && window.matchMedia('(min-width: 1024px)')) || { matches: false, addEventListener() {}, addListener() {} };
    const pdfFrame1 = document.getElementById('pdf-frame-1');
    const pdfFrame2 = document.getElementById('pdf-frame-2');
    const liveEmpty = document.getElementById('live-pdf-preview-empty');
    const livePreviewFrame = document.querySelector('.live-preview-frame');
    const livePreviewWrap = document.querySelector('.live-preview-wrap');
    const kreatorCard = document.querySelector('.kreator-layout > .card');
    const kreatorLayout = document.querySelector('.kreator-layout');
    const pdfLightbox = document.getElementById('pdf-lightbox');
    const pdfLightboxBackdrop = document.getElementById('pdf-lightbox-backdrop');
    const pdfLightboxClose = document.getElementById('pdf-lightbox-close');
    const pdfLightboxFrame = document.getElementById('pdf-lightbox-frame');
    let activeFrame = pdfFrame1;
    let bufferFrame = pdfFrame2;
    let activeUrl = '';
    let pendingUrl = '';
    let livePodgladAbort = null;

    const PDF_VIEW_PARAMS = '#toolbar=0&navpanes=0&scrollbar=0&statusbar=0&messages=0&view=FitH&zoom=page-fit';
    const PDF_FULLSCREEN_PARAMS = '#toolbar=0&navpanes=0&scrollbar=0&statusbar=0&messages=0&view=Fit';

    function ustawHasPdf(stan) {
      if (livePreviewFrame) livePreviewFrame.classList.toggle('has-pdf', stan);
    }

    let previewHeightSyncRaf = 0;
    const A4_PREVIEW_RATIO = 1.414;

    function applyPreviewSize() {
      if (!livePreviewFrame || !kreatorCard || !livePreviewWrap) return;
      if (!livePodgladMql.matches) {
        livePreviewFrame.style.removeProperty('width');
        livePreviewFrame.style.removeProperty('height');
        return;
      }
      const cardH = kreatorCard.getBoundingClientRect().height;
      const colW = livePreviewWrap.clientWidth;
      if (cardH <= 0 || colW <= 0) return;
      const h = Math.min(cardH, colW * A4_PREVIEW_RATIO);
      const w = h / A4_PREVIEW_RATIO;
      livePreviewFrame.style.width = Math.round(w) + 'px';
      livePreviewFrame.style.height = Math.round(h) + 'px';
    }

    function synchronizujWysokoscPodgladu() {
      cancelAnimationFrame(previewHeightSyncRaf);
      previewHeightSyncRaf = requestAnimationFrame(() => {
        applyPreviewSize();
        requestAnimationFrame(applyPreviewSize);
      });
    }

    if (window.ResizeObserver) {
      const previewSizeRo = new ResizeObserver(synchronizujWysokoscPodgladu);
      if (kreatorCard) previewSizeRo.observe(kreatorCard);
      if (livePreviewWrap) previewSizeRo.observe(livePreviewWrap);
      if (kreatorLayout) previewSizeRo.observe(kreatorLayout);
    }
    window.addEventListener('resize', synchronizujWysokoscPodgladu);
    if (window.visualViewport) {
      window.visualViewport.addEventListener('resize', synchronizujWysokoscPodgladu);
      window.visualViewport.addEventListener('scroll', synchronizujWysokoscPodgladu);
    }
    [
      '(min-width: 1024px) and (max-width: 1365px)',
      '(min-width: 1366px) and (max-width: 1919px)',
      '(min-width: 1920px) and (max-width: 2559px)',
      '(min-width: 2560px)',
    ].forEach((query) => {
      const mql = window.matchMedia(query);
      if (mql.addEventListener) mql.addEventListener('change', synchronizujWysokoscPodgladu);
      else if (mql.addListener) mql.addListener(synchronizujWysokoscPodgladu);
    });
    synchronizujWysokoscPodgladu();

    function pokazPlaceholderLive() {
      if (!activeFrame || !bufferFrame || !liveEmpty) return;
      activeFrame.classList.remove('active');
      bufferFrame.classList.remove('active');
      if (pendingUrl) {
        URL.revokeObjectURL(pendingUrl);
        pendingUrl = '';
      }
      liveEmpty.hidden = false;
      ustawHasPdf(false);
      synchronizujSrcLightbox();
      if (pdfLightbox && pdfLightbox.classList.contains('is-open')) {
        zamknijPdfLightbox();
      }
    }

    function docelowySrcLightbox() {
      return activeUrl ? activeUrl + PDF_VIEW_PARAMS : '';
    }

    function synchronizujSrcLightbox() {
      if (!pdfLightboxFrame) return;
      const docelowy = docelowySrcLightbox();
      if (!docelowy) {
        pdfLightboxFrame.removeAttribute('src');
        return;
      }
      const biezacy = pdfLightboxFrame.getAttribute('src') || '';
      if (biezacy !== docelowy) {
        pdfLightboxFrame.src = docelowy;
      }
    }

    function aktywujBuforPdf(oczekiwanyUrl) {
      if (!bufferFrame || !pendingUrl || pendingUrl !== oczekiwanyUrl) return;

      bufferFrame.classList.add('active');
      activeFrame.classList.remove('active');
      if (liveEmpty) liveEmpty.hidden = true;
      ustawHasPdf(true);

      const oldUrl = activeUrl;
      activeUrl = pendingUrl;
      pendingUrl = '';

      synchronizujSrcLightbox();

      const tmp = activeFrame;
      activeFrame = bufferFrame;
      bufferFrame = tmp;

      if (oldUrl) {
        setTimeout(() => URL.revokeObjectURL(oldUrl), 800);
      }
    }

    function onFrameLoad(frame) {
      if (frame !== bufferFrame || !pendingUrl) return;
      const url = pendingUrl;
      requestAnimationFrame(() => {
        requestAnimationFrame(() => aktywujBuforPdf(url));
      });
    }

    if (pdfFrame1) pdfFrame1.addEventListener('load', () => onFrameLoad(pdfFrame1));
    if (pdfFrame2) pdfFrame2.addEventListener('load', () => onFrameLoad(pdfFrame2));

    function zaladujPdfDoBufora(blob) {
      if (!activeFrame || !bufferFrame) return;
      const nowyUrl = URL.createObjectURL(blob);
      if (pendingUrl) {
        URL.revokeObjectURL(pendingUrl);
      }
      pendingUrl = nowyUrl;
      bufferFrame.src = nowyUrl + PDF_VIEW_PARAMS;
      setTimeout(() => aktywujBuforPdf(nowyUrl), 1200);
    }

    let pdfLightboxZrodlo = '';

    function otworzPdfLightbox(zrodloUrl) {
      const url = zrodloUrl || activeUrl;
      if (!url || !pdfLightbox || !pdfLightboxFrame) return;
      pdfLightboxZrodlo = zrodloUrl ? 'modal' : 'live';
      const params = pdfLightboxZrodlo === 'modal'
        && window.matchMedia('(max-width: 1023px)').matches
        ? PDF_VIEW_PARAMS
        : (pdfLightboxZrodlo === 'modal' ? PDF_FULLSCREEN_PARAMS : PDF_VIEW_PARAMS);
      const docelowy = url + params;
      const biezacy = pdfLightboxFrame.getAttribute('src') || '';
      if (biezacy !== docelowy) {
        pdfLightboxFrame.src = docelowy;
      }
      pdfLightbox.classList.toggle('is-from-modal', pdfLightboxZrodlo === 'modal');
      pdfLightbox.classList.remove('is-settled');
      pdfLightbox.hidden = false;
      pdfLightbox.setAttribute('aria-hidden', 'false');
      const modalPdf = document.getElementById('pdf-modal');
      if (!modalPdf || modalPdf.hidden) {
        document.body.style.overflow = 'hidden';
      }
      requestAnimationFrame(() => {
        pdfLightbox.classList.add('is-open');
        requestAnimationFrame(() => {
          window.setTimeout(() => {
            if (pdfLightbox.classList.contains('is-open')) {
              pdfLightbox.classList.add('is-settled');
            }
          }, 220);
        });
      });
    }

    function zamknijPdfLightbox() {
      if (!pdfLightbox) return;
      const zrodlo = pdfLightboxZrodlo;
      pdfLightbox.classList.remove('is-open', 'is-settled');
      pdfLightbox.setAttribute('aria-hidden', 'true');
      pdfLightboxZrodlo = '';
      pdfLightbox.classList.remove('is-from-modal');
      window.setTimeout(() => {
        if (pdfLightbox && !pdfLightbox.classList.contains('is-open')) {
          pdfLightbox.hidden = true;
        }
      }, 220);
      const modalPdf = document.getElementById('pdf-modal');
      if (zrodlo === 'modal' && modalPdf && !modalPdf.hidden) {
        document.body.style.overflow = 'hidden';
        return;
      }
      document.body.style.overflow = '';
    }

    if (livePreviewFrame) {
      livePreviewFrame.addEventListener('click', () => {
        if (livePreviewFrame.classList.contains('has-pdf')) {
          otworzPdfLightbox();
        }
      });
    }
    if (pdfLightboxBackdrop) pdfLightboxBackdrop.addEventListener('click', zamknijPdfLightbox);
    if (pdfLightboxClose) pdfLightboxClose.addEventListener('click', zamknijPdfLightbox);
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && pdfLightbox && pdfLightbox.classList.contains('is-open')) {
        zamknijPdfLightbox();
      }
    });

    function aktualizujTekstPrzyciskuGeneruj() {
      if (!btnGeneruj) return;
      btnGeneruj.textContent = livePodgladMql.matches ? 'Pobierz PDF' : 'Wygeneruj wycenę';
    }
    aktualizujTekstPrzyciskuGeneruj();

    async function aktualizujLivePodglad() {
      if (!livePodgladMql.matches) return;
      if (!activeFrame || !bufferFrame) return;

      const { gotowy, payload } = budujPayloadZFormularza();
      if (!gotowy) {
        pokazPlaceholderLive();
        if (livePodgladAbort) { livePodgladAbort.abort(); livePodgladAbort = null; }
        return;
      }

      if (livePodgladAbort) livePodgladAbort.abort();
      const ctrl = new AbortController();
      livePodgladAbort = ctrl;

      try {
        const res = await fetch('/quote', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
          signal: ctrl.signal,
        });
        if (!res.ok) return;
        const blob = await res.blob();
        if (ctrl.signal.aborted) return;

        zaladujPdfDoBufora(blob);
      } catch (err) {
        if (err && err.name === 'AbortError') return;
      } finally {
        if (livePodgladAbort === ctrl) livePodgladAbort = null;
      }
    }

    const btnUdostepnijPdfDesktop = document.getElementById('btn-udostepnij-pdf-desktop');

    function aktualizujWidocznoscShareDesktop() {
      if (!btnUdostepnijPdfDesktop) return;
      btnUdostepnijPdfDesktop.hidden = !(shareSupported && livePodgladMql.matches);
    }
    aktualizujWidocznoscShareDesktop();

    const livePodgladMqlHandler = () => {
      aktualizujTekstPrzyciskuGeneruj();
      aktualizujWidocznoscShareDesktop();
      synchronizujWysokoscPodgladu();
      if (livePodgladMql.matches) {
        aktualizujLivePodglad();
      } else if (livePodgladAbort) {
        livePodgladAbort.abort();
        livePodgladAbort = null;
      }
    };
    if (livePodgladMql.addEventListener) livePodgladMql.addEventListener('change', livePodgladMqlHandler);
    else if (livePodgladMql.addListener) livePodgladMql.addListener(livePodgladMqlHandler);

    if (shareSupported && btnUdostepnijPdfDesktop) {
      btnUdostepnijPdfDesktop.addEventListener('click', async () => {
        ukryjKomunikat();

        const cfg = wczytajConfig();
        if (!String(cfg.nazwa_firmy || '').trim()) {
          pokazKomunikat('Uzupełnij nazwę firmy w zakładce Moja firma.', 'error');
          przejdzDoMojaFirma();
          return;
        }

        const { payload, koszty } = budujPayloadZFormularza();
        const originalLabel = btnUdostepnijPdfDesktop.querySelector('span');
        const originalText = originalLabel ? originalLabel.textContent : '';
        btnUdostepnijPdfDesktop.disabled = true;
        if (originalLabel) originalLabel.textContent = 'Przygotowuję…';

        try {
          const res = await fetch('/quote', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
          });

          if (!res.ok) {
            const tekst = await res.text();
            throw new Error(tekst || `Błąd ${res.status}`);
          }

          const blob = await res.blob();
          const nazwaPliku = `wycena-${payload.klient.replace(/[^a-z0-9-_]+/gi, '_') || 'dokument'}.pdf`;
          const file = new File([blob], nazwaPliku, { type: 'application/pdf' });

          if (!navigator.canShare({ files: [file] })) {
            pokazKomunikat('Ta przeglądarka nie wspiera wysyłania plików PDF. Użyj „Pobierz PDF" i załącz ręcznie.', 'error');
            return;
          }

          await navigator.share({
            files: [file],
            title: 'Wycena PDF',
            text: 'Wycena w załączniku.',
          });

          if (livePodgladMql.matches) {
            zaladujPdfDoBufora(blob);
          }
          const _pendingTok1 = window._pendingLinkToken || null;
          window._pendingLinkToken = null;
          dodajDoHistorii(payload, koszty, _pendingTok1);
          inkrementujNumeracje();
          renderStatystyki();
          const numerEl = document.getElementById('numer_oferty');
          if (numerEl) numerEl.value = nastepnyNumerOferty();
          aktualnyPayloadEmail = payload;
          pokazKomunikat('Wycena wysłana do klienta i zapisana w historii.', 'success');
        } catch (err) {
          if (err && err.name === 'AbortError') return;
          pokazKomunikat('Nie udało się wysłać pliku: ' + err.message, 'error');
        } finally {
          btnUdostepnijPdfDesktop.disabled = false;
          if (originalLabel) originalLabel.textContent = originalText || 'Wyślij klientowi';
        }
      });
    }

    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      ukryjKomunikat();

      const cfg = wczytajConfig();
      if (!String(cfg.nazwa_firmy || '').trim()) {
        pokazKomunikat('Uzupełnij nazwę firmy w zakładce Moja firma.', 'error');
        przejdzDoMojaFirma();
        return;
      }

      const { payload, koszty } = budujPayloadZFormularza();
      const desktopTryb = livePodgladMql.matches;

      btnGeneruj.disabled = true;
      btnGeneruj.textContent = desktopTryb ? 'Pobieranie...' : 'Generowanie...';

      try {
        const res = await fetch('/quote', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });

        if (!res.ok) {
          const tekst = await res.text();
          throw new Error(tekst || `Błąd ${res.status}`);
        }

        const blob = await res.blob();
        const nazwaPliku = `wycena-${payload.klient.replace(/[^a-z0-9-_]+/gi, '_') || 'dokument'}.pdf`;

        trackEvent('pdf_generated');

        if (desktopTryb) {
          zaladujPdfDoBufora(blob);

          const dlUrl = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = dlUrl;
          a.download = nazwaPliku;
          document.body.appendChild(a);
          a.click();
          a.remove();
          setTimeout(() => URL.revokeObjectURL(dlUrl), 10000);

          const _pendingTok2 = window._pendingLinkToken || null;
          window._pendingLinkToken = null;
          dodajDoHistorii(payload, koszty, _pendingTok2);
          inkrementujNumeracje();
          renderStatystyki();
          const numerEl = document.getElementById('numer_oferty');
          if (numerEl) numerEl.value = nastepnyNumerOferty();
          aktualnyPayloadEmail = payload;

          pokazKomunikat('PDF pobrany. Wycena zapisana w historii.', 'success');
        } else {
          const url = URL.createObjectURL(blob);
          pokazPodgladPdf(url, nazwaPliku, payload, blob, koszty);
          pokazKomunikat('PDF wygenerowany. Otwórz podgląd, aby pobrać plik.', 'success');
        }
      } catch (err) {
        pokazKomunikat('Nie udało się wygenerować PDF: ' + err.message, 'error');
      } finally {
        btnGeneruj.disabled = false;
        aktualizujTekstPrzyciskuGeneruj();
      }
    });

    const pdfModal = document.getElementById('pdf-modal');
    const pdfModalBackdrop = document.getElementById('pdf-modal-backdrop');
    const pdfPreview = document.getElementById('pdf-preview');
    const btnPobierzPdf = document.getElementById('btn-pobierz-pdf');
    const btnDrukujPdf = document.getElementById('btn-drukuj-pdf');
    const btnPrzygotujEmail = document.getElementById('btn-przygotuj-email');
    const btnZamknijModal = document.getElementById('btn-zamknij-modal');
    const pdfModalMobileMql = window.matchMedia('(max-width: 1023px)');

    let aktualnyPodgladUrl = '';
    let aktualnaNazwaPliku = '';
    let aktualnyPayload = null;
    let aktualnyPayloadEmail = null;
    let aktualnyBlobPdf = null;
    let aktualneKoszty = null;

    function ukryjPdfModal() {
      if (pdfLightbox && pdfLightbox.classList.contains('is-open') && pdfLightboxZrodlo === 'modal') {
        zamknijPdfLightbox();
      }
      pdfModal.hidden = true;
      pdfModal.setAttribute('aria-hidden', 'true');
      pdfModal.classList.remove('is-open');
      pdfPreview.removeAttribute('src');
      document.body.style.overflow = '';
      if (aktualnyPodgladUrl) {
        URL.revokeObjectURL(aktualnyPodgladUrl);
        aktualnyPodgladUrl = '';
      }
      aktualnaNazwaPliku = '';
      aktualnyPayload = null;
      aktualnyPayloadEmail = null;
      aktualnyBlobPdf = null;
      aktualneKoszty = null;
    }

    function pokazPodgladPdf(url, nazwaPliku, payload, blob, koszty) {
      if (aktualnyPodgladUrl) {
        URL.revokeObjectURL(aktualnyPodgladUrl);
      }
      aktualnyPodgladUrl = url;
      aktualnaNazwaPliku = nazwaPliku;
      aktualnyPayload = payload || null;
      aktualnyPayloadEmail = payload || null;
      aktualnyBlobPdf = blob || null;
      aktualneKoszty = Array.isArray(koszty) ? koszty.slice() : null;
      pdfPreview.src = url + PDF_VIEW_PARAMS;
      pdfModal.hidden = false;
      pdfModal.setAttribute('aria-hidden', 'false');
      document.body.style.overflow = 'hidden';
      if (pdfModalMobileMql.matches) {
        pdfModal.classList.remove('is-open');
        if (animacjeWlaczone()) {
          requestAnimationFrame(() => pdfModal.classList.add('is-open'));
        } else {
          pdfModal.classList.add('is-open');
        }
      }
    }

    function zamknijPodgladPdf() {
      if (pdfModal.hidden) return;
      if (
        pdfModalMobileMql.matches &&
        pdfModal.classList.contains('is-open') &&
        animacjeWlaczone()
      ) {
        pdfModal.classList.remove('is-open');
        const card = pdfModal.querySelector('.pdf-preview-sheet');
        if (!card) {
          ukryjPdfModal();
          return;
        }
        let zamkniete = false;
        const finalize = () => {
          if (zamkniete) return;
          zamkniete = true;
          ukryjPdfModal();
        };
        const done = (e) => {
          if (e.target !== card || e.propertyName !== 'transform') return;
          card.removeEventListener('transitionend', done);
          finalize();
        };
        card.addEventListener('transitionend', done);
        window.setTimeout(finalize, 320);
        return;
      }
      ukryjPdfModal();
    }

    btnZamknijModal.addEventListener('click', zamknijPodgladPdf);
    pdfModalBackdrop.addEventListener('click', zamknijPodgladPdf);
    const pdfPreviewTap = document.getElementById('pdf-preview-tap');
    if (pdfPreviewTap) {
      const otworzPelnyPodgladZModala = () => {
        if (!aktualnyPodgladUrl) return;
        otworzPdfLightbox(aktualnyPodgladUrl);
      };
      pdfPreviewTap.addEventListener('click', otworzPelnyPodgladZModala);
      pdfPreviewTap.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          otworzPelnyPodgladZModala();
        }
      });
    }
    pdfModalMobileMql.addEventListener('change', () => {
      if (!pdfModal || pdfModal.hidden || pdfModalMobileMql.matches) return;
      pdfModal.classList.remove('is-open');
    });
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && !pdfModal.hidden) {
        if (pdfLightbox && pdfLightbox.classList.contains('is-open')) return;
        zamknijPodgladPdf();
      }
    });

    btnPobierzPdf.addEventListener('click', () => {
      if (!aktualnyPodgladUrl) return;
      const a = document.createElement('a');
      a.href = aktualnyPodgladUrl;
      a.download = aktualnaNazwaPliku;
      document.body.appendChild(a);
      a.click();
      a.remove();
      if (aktualnyPayload) {
        dodajDoHistorii(aktualnyPayload, aktualneKoszty);
        inkrementujNumeracje();
        renderStatystyki();
        aktualnyPayload = null;
        aktualneKoszty = null;
      }
    });

    const btnUdostepnijPdf = document.getElementById('btn-udostepnij-pdf');

    if (shareSupported && btnUdostepnijPdf) {
      btnUdostepnijPdf.hidden = false;
      btnUdostepnijPdf.addEventListener('click', async () => {
        if (!aktualnyBlobPdf) return;
        const file = new File([aktualnyBlobPdf], aktualnaNazwaPliku || 'Wycena.pdf', { type: 'application/pdf' });
          if (!navigator.canShare({ files: [file] })) {
            pokazKomunikat('Ta przeglądarka nie wspiera wysyłania plików PDF.', 'error');
            return;
          }
          try {
            await navigator.share({
              files: [file],
              title: 'Wycena PDF',
              text: 'Wycena w załączniku.',
            });
            if (aktualnyPayload) {
              dodajDoHistorii(aktualnyPayload, aktualneKoszty);
              inkrementujNumeracje();
              renderStatystyki();
              aktualnyPayload = null;
              aktualneKoszty = null;
            }
            pokazKomunikat('Wycena wysłana do klienta.', 'success');
          } catch (err) {
            if (err && err.name === 'AbortError') return;
            pokazKomunikat('Nie udało się wysłać pliku: ' + err.message, 'error');
          }
      });
    }

    btnDrukujPdf.addEventListener('click', () => {
      if (!aktualnyPodgladUrl) return;
      try {
        pdfPreview.contentWindow.focus();
        pdfPreview.contentWindow.print();
      } catch (e) {
        pokazKomunikat('Nie udało się otworzyć dialogu drukowania. Pobierz PDF i wydrukuj z czytnika.', 'error');
      }
    });

    btnPrzygotujEmail.addEventListener('click', () => {
      if (!aktualnyPayloadEmail) {
        pokazKomunikat('Najpierw wygeneruj wycenę, aby przygotować e-mail.', 'error');
        return;
      }
      const p = aktualnyPayloadEmail;
      const numer = String(p.numer_oferty || '').trim();
      const firma = String(p.nazwa_firmy || '').trim();
      const sumaNum = obliczSumePozycji(p.pozycje);
      const sumaTxt = (Number.isFinite(sumaNum) ? sumaNum : 0).toFixed(2).replace('.', ',');

      const numerTxt = numer || 'bez numeru';
      const subject = firma
        ? `Oferta ${numerTxt} - ${firma}`
        : `Oferta ${numerTxt}`;
      const podpis = firma ? `Pozdrawiam,\n${firma}` : 'Pozdrawiam';
      const body = `Dzień dobry,\n\nw załączniku przesyłam ofertę ${numerTxt} na kwotę ${sumaTxt} zł. W razie pytań jestem do dyspozycji.\n\n${podpis}`;

      const mailto = 'mailto:?subject=' + encodeURIComponent(subject) + '&body=' + encodeURIComponent(body);
      window.location.href = mailto;
    });

    const STORAGE_KEY_CONFIG = 'sumit_config';
    const STORAGE_KEY_CONFIG_LEGACY = 'sumit_dane_firmy';
    const POLA_CONFIG = ['nazwa_firmy', 'nip', 'adres', 'miasto', 'telefon', 'email', 'numer_konta'];
    const MAX_LOGO_BYTES = 200 * 1024;

    const appSettingsModal = document.getElementById('app-settings-modal');
    const appSettingsBackdrop = document.getElementById('app-settings-modal-backdrop');
    const appSettingsMobileMql = window.matchMedia('(max-width: 1023px)');
    const btnAppSettingsZamknij = document.getElementById('btn-app-settings-zamknij');
    const btnSettings = document.getElementById('btn-settings');
    const btnSettingsMobile = document.getElementById('btn-settings-mobile');
    const btnCfgWyczysc = document.getElementById('btn-cfg-wyczysc');
    const btnCfgZapisz = document.getElementById('btn-cfg-zapisz');
    const settingsMessage = document.getElementById('settings-message');
    const settingsForm = document.getElementById('settings-form');

    const cfgLogoInput = document.getElementById('cfg-logo-input');
    const btnCfgDodajLogo = document.getElementById('btn-cfg-dodaj-logo');
    const btnCfgUsunLogo = document.getElementById('btn-cfg-usun-logo');
    const cfgLogoPreviewImg = document.getElementById('cfg-logo-preview');
    const firmaLogoPlaceholder = document.getElementById('firma-logo-placeholder');
    const btnCfgPobierzNip = document.getElementById('btn-cfg-pobierz-nip');
    const firmaSaveStatus = document.getElementById('firma-save-status');
    let cfgNipFetchController = null;
    let firmaSaveStatusTimer = null;

    let cfgLogoBase64 = '';

    function ustawLiniePodgladuFirmy(el, tekst) {
      if (!el) return;
      const val = typeof tekst === 'string' ? tekst.trim() : '';
      if (val) {
        el.textContent = val;
        el.hidden = false;
      } else {
        el.textContent = '';
        el.hidden = true;
      }
    }

    function odswiezPodgladFirmy() {
      const cfg = zbierzConfigZFormularza();
      const previewLogo = document.getElementById('firma-preview-logo');
      const previewNazwa = document.getElementById('firma-preview-nazwa');
      const previewNip = document.getElementById('firma-preview-nip');
      const previewAdres = document.getElementById('firma-preview-adres');
      const previewMiasto = document.getElementById('firma-preview-miasto');
      const previewTelefon = document.getElementById('firma-preview-telefon');
      const previewEmail = document.getElementById('firma-preview-email');
      const statusQr = document.getElementById('firma-status-qr');
      const statusLogo = document.getElementById('firma-status-logo');

      if (previewLogo) {
        if (cfg.logo_base64) {
          previewLogo.src = cfg.logo_base64;
          previewLogo.hidden = false;
        } else {
          previewLogo.removeAttribute('src');
          previewLogo.hidden = true;
        }
      }

      if (previewNazwa) {
        const nazwa = cfg.nazwa_firmy.trim();
        previewNazwa.textContent = nazwa || 'Nazwa firmy';
        previewNazwa.classList.toggle('is-empty', !nazwa);
      }

      ustawLiniePodgladuFirmy(previewNip, cfg.nip.trim() ? 'NIP: ' + cfg.nip.trim() : '');
      ustawLiniePodgladuFirmy(previewAdres, cfg.adres);
      ustawLiniePodgladuFirmy(previewMiasto, cfg.miasto);
      ustawLiniePodgladuFirmy(previewTelefon, cfg.telefon.trim() ? 'tel: ' + cfg.telefon.trim() : '');
      ustawLiniePodgladuFirmy(previewEmail, cfg.email.trim() ? cfg.email.trim() : '');

      const cyfryKonta = (cfg.numer_konta || '').replace(/\D+/g, '');
      if (statusQr) {
        if (cyfryKonta.length === 26) {
          statusQr.textContent = 'QR na PDF';
          statusQr.dataset.state = 'ok';
        } else if (cyfryKonta.length > 0) {
          statusQr.textContent = 'NRB: ' + cyfryKonta.length + '/26';
          statusQr.dataset.state = 'warn';
        } else {
          statusQr.textContent = 'Bez QR';
          statusQr.dataset.state = 'warn';
        }
      }

      if (statusLogo) {
        if (cfg.logo_base64) {
          statusLogo.textContent = 'Logo OK';
          statusLogo.dataset.state = 'ok';
        } else {
          statusLogo.textContent = 'Bez logo';
          statusLogo.dataset.state = 'warn';
        }
      }
    }

    function pokazStatusZapisuFirmy() {
      if (!firmaSaveStatus) return;
      firmaSaveStatus.textContent = 'Zapisano';
      if (firmaSaveStatusTimer) clearTimeout(firmaSaveStatusTimer);
      firmaSaveStatusTimer = setTimeout(() => {
        if (firmaSaveStatus) firmaSaveStatus.textContent = '';
      }, 2000);
    }

    function ustawStanLadowaniaCfgNIP(loading) {
      if (!btnCfgPobierzNip) return;
      btnCfgPobierzNip.disabled = loading;
      btnCfgPobierzNip.setAttribute('aria-busy', loading ? 'true' : 'false');
      const label = btnCfgPobierzNip.querySelector('.btn-link-nip-label');
      if (label) label.textContent = loading ? 'Pobieram…' : 'Pobierz z MF';
    }

    async function pobierzFirmePoNIP() {
      const nipEl = document.getElementById('cfg-nip');
      if (!nipEl) return;

      const nip = oczyscNIPWejscie(nipEl.value);
      if (nip.length !== 10) {
        pokazKomunikatCfg('NIP musi mieć dokładnie 10 cyfr.', 'error');
        nipEl.focus();
        return;
      }

      ustawStanLadowaniaCfgNIP(true);
      if (cfgNipFetchController) cfgNipFetchController.abort();
      cfgNipFetchController = new AbortController();

      try {
        const resp = await fetch('/api/nip?nip=' + encodeURIComponent(nip), {
          headers: { Accept: 'application/json' },
          signal: cfgNipFetchController.signal,
        });
        let dane = null;
        try { dane = await resp.json(); } catch (e) { dane = null; }

        if (!resp.ok) {
          const msg = (dane && dane.error) ? dane.error : 'Nie udało się pobrać danych z Białej Listy MF.';
          pokazKomunikatCfg(msg, 'error');
          return;
        }

        const nazwaEl = document.getElementById('cfg-nazwa_firmy');
        const adresEl = document.getElementById('cfg-adres');
        if (nazwaEl && dane && dane.nazwa) nazwaEl.value = String(dane.nazwa);
        if (adresEl && dane && dane.adres) adresEl.value = String(dane.adres);
        if (dane && dane.nip) {
          const s = String(dane.nip).replace(/\D/g, '');
          if (s.length === 10) {
            nipEl.value = s.slice(0, 3) + '-' + s.slice(3, 6) + '-' + s.slice(6, 8) + '-' + s.slice(8, 10);
          }
        }

        ukryjKomunikatCfg();
        zapiszCfgZFormularza();
        odswiezPodgladFirmy();
      } catch (err) {
        if (err && err.name === 'AbortError') return;
        pokazKomunikatCfg('Błąd sieci: nie udało się połączyć z serwerem.', 'error');
      } finally {
        cfgNipFetchController = null;
        ustawStanLadowaniaCfgNIP(false);
      }
    }

    function pokazKomunikatCfg(tekst, typ) {
      if (!settingsMessage) return;
      settingsMessage.textContent = tekst;
      settingsMessage.className = 'message ' + (typ || '');
    }

    function ukryjKomunikatCfg() {
      if (!settingsMessage) return;
      settingsMessage.textContent = '';
      settingsMessage.className = 'message';
    }

    function migrujStaryConfig() {
      try {
        if (localStorage.getItem(STORAGE_KEY_CONFIG)) return;
        const raw = localStorage.getItem(STORAGE_KEY_CONFIG_LEGACY);
        if (!raw) return;
        const stare = JSON.parse(raw);
        if (!stare || typeof stare !== 'object') return;
        const nowe = {};
        POLA_CONFIG.forEach(k => {
          if (typeof stare[k] === 'string') nowe[k] = stare[k];
        });
        localStorage.setItem(STORAGE_KEY_CONFIG, JSON.stringify(nowe));
        localStorage.removeItem(STORAGE_KEY_CONFIG_LEGACY);
      } catch (e) {}
    }

    function wczytajConfig() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY_CONFIG);
        if (!raw) return {};
        const dane = JSON.parse(raw);
        if (!dane || typeof dane !== 'object') return {};
        return dane;
      } catch (e) {
        return {};
      }
    }

    function zapiszConfig(cfg) {
      try {
        localStorage.setItem(STORAGE_KEY_CONFIG, JSON.stringify(cfg));
      } catch (e) {
        pokazKomunikatCfg('Nie udało się zapisać ustawień (brak miejsca w przeglądarce).', 'error');
      }
    }

    function ustawPodgladLogo(dataUrl) {
      if (dataUrl) {
        cfgLogoBase64 = dataUrl;
        if (cfgLogoPreviewImg) {
          cfgLogoPreviewImg.src = dataUrl;
          cfgLogoPreviewImg.hidden = false;
        }
        if (firmaLogoPlaceholder) firmaLogoPlaceholder.hidden = true;
        if (btnCfgUsunLogo) btnCfgUsunLogo.hidden = false;
        if (btnCfgDodajLogo) btnCfgDodajLogo.setAttribute('aria-label', 'Zmień logo firmy (PNG lub JPG, maks. 200 KB)');
      } else {
        cfgLogoBase64 = '';
        if (cfgLogoPreviewImg) {
          cfgLogoPreviewImg.removeAttribute('src');
          cfgLogoPreviewImg.hidden = true;
        }
        if (firmaLogoPlaceholder) firmaLogoPlaceholder.hidden = false;
        if (btnCfgUsunLogo) btnCfgUsunLogo.hidden = true;
        if (btnCfgDodajLogo) btnCfgDodajLogo.setAttribute('aria-label', 'Wybierz logo firmy (PNG lub JPG, maks. 200 KB)');
      }
      odswiezPodgladFirmy();
    }

    function wypelnijFormularzCfg() {
      const cfg = wczytajConfig();
      POLA_CONFIG.forEach(k => {
        const el = document.getElementById('cfg-' + k);
        if (el) el.value = typeof cfg[k] === 'string' ? cfg[k] : '';
      });
      ustawPodgladLogo(typeof cfg.logo_base64 === 'string' ? cfg.logo_base64 : '');
      ukryjKomunikatCfg();
      odswiezPodgladFirmy();
    }

    function zbierzConfigZFormularza() {
      const cfg = {};
      POLA_CONFIG.forEach(k => {
        const el = document.getElementById('cfg-' + k);
        cfg[k] = el ? el.value.trim() : '';
      });
      cfg.logo_base64 = cfgLogoBase64 || '';
      return cfg;
    }

    function walidujNumerKonta() {
      const el = document.getElementById('cfg-numer_konta');
      if (!el) return true;
      const raw = el.value.trim().toUpperCase().replace(/^PL/, '');
      const cyfry = raw.replace(/\D+/g, '');
      if (cyfry.length === 0) return true;
      if (cyfry.length !== 26) {
        pokazKomunikatCfg('Numer konta musi mieć dokładnie 26 cyfr (obecnie ' + cyfry.length + ').', 'error');
        el.focus();
        return false;
      }
      return true;
    }

    function zapiszCfgZFormularza() {
      zapiszConfig(zbierzConfigZFormularza());
      pokazStatusZapisuFirmy();
      odswiezPodgladFirmy();
    }

    if (btnCfgDodajLogo) btnCfgDodajLogo.addEventListener('click', () => cfgLogoInput.click());
    if (btnCfgPobierzNip) btnCfgPobierzNip.addEventListener('click', pobierzFirmePoNIP);

    if (cfgLogoInput) cfgLogoInput.addEventListener('change', () => {
      const file = cfgLogoInput.files && cfgLogoInput.files[0];
      if (!file) return;
      if (!/^image\/(png|jpe?g)$/i.test(file.type)) {
        pokazKomunikatCfg('Wybierz plik PNG lub JPG.', 'error');
        cfgLogoInput.value = '';
        return;
      }
      const reader = new FileReader();
      reader.onload = () => {
        const dataUrl = String(reader.result || '');
        if (dataUrl.length > MAX_LOGO_BYTES) {
          const kb = Math.round(dataUrl.length / 1024);
          pokazKomunikatCfg('Logo musi być ≤ 200 KB po konwersji (aktualnie ~' + kb + ' KB). Wybierz mniejszy plik.', 'error');
          cfgLogoInput.value = '';
          return;
        }
        ustawPodgladLogo(dataUrl);
        ukryjKomunikatCfg();
        zapiszCfgZFormularza();
      };
      reader.onerror = () => {
        pokazKomunikatCfg('Nie udało się wczytać pliku logo.', 'error');
        cfgLogoInput.value = '';
      };
      reader.readAsDataURL(file);
    });

    if (btnCfgUsunLogo) btnCfgUsunLogo.addEventListener('click', () => {
      ustawPodgladLogo('');
      if (cfgLogoInput) cfgLogoInput.value = '';
      zapiszCfgZFormularza();
    });

    if (btnCfgWyczysc) btnCfgWyczysc.addEventListener('click', () => {
      POLA_CONFIG.forEach(k => {
        const el = document.getElementById('cfg-' + k);
        if (el) el.value = '';
      });
      ustawPodgladLogo('');
      cfgLogoInput.value = '';
      try { localStorage.removeItem(STORAGE_KEY_CONFIG); } catch (e) {}
      pokazKomunikatCfg('Wyczyszczono ustawienia firmy.', 'success');
    });

    if (btnCfgZapisz) {
      btnCfgZapisz.addEventListener('click', () => {
        if (!walidujNumerKonta()) return;
        zapiszCfgZFormularza();
      });
    }

    if (settingsForm) {
      settingsForm.addEventListener('change', () => {
        if (!walidujNumerKonta()) return;
        zapiszCfgZFormularza();
      });
      settingsForm.addEventListener('input', () => {
        odswiezPodgladFirmy();
      });
    }

    let aktywujWidok = null;

    function widokFirmaAktywny() {
      const firma = document.getElementById('view-firma');
      return !!(firma && !firma.classList.contains('hidden') && !firma.hasAttribute('hidden'));
    }

    function zapiszFirmeJesliTrzeba() {
      if (!widokFirmaAktywny()) return true;
      if (!walidujNumerKonta()) return false;
      zapiszCfgZFormularza();
      return true;
    }

    function przejdzDoMojaFirma() {
      const tab = document.getElementById('tab-ustawienia');
      if (tab) tab.click();
    }

    function odswiezChipyStatsOkresApp() {
      const wrap = document.getElementById('app-settings-stats-okres');
      if (!wrap) return;
      const okres = wczytajOkresStat();
      wrap.querySelectorAll('.chip[data-stats-okres]').forEach((chip) => {
        const active = chip.getAttribute('data-stats-okres') === okres;
        chip.classList.toggle('is-active', active);
        chip.setAttribute('aria-pressed', active ? 'true' : 'false');
      });
    }

    function odswiezWszystkieChipyAppSettings() {
      odswiezChipyMotywuApp();
      odswiezChipyWaznosciApp();
      odswiezChipyStatsOkresApp();
      odswiezChipyDocTypeApp();
      odswiezChipyVatApp();
    }

    function ukryjAppSettingsModal() {
      if (!appSettingsModal) return;
      appSettingsModal.hidden = true;
      appSettingsModal.setAttribute('aria-hidden', 'true');
      appSettingsModal.classList.remove('is-open');
      document.body.style.overflow = '';
    }

    function otworzAppSettings() {
      if (!appSettingsModal) return;
      odswiezWszystkieChipyAppSettings();
      appSettingsModal.hidden = false;
      appSettingsModal.setAttribute('aria-hidden', 'false');
      document.body.style.overflow = 'hidden';
      if (appSettingsMobileMql.matches) {
        appSettingsModal.classList.remove('is-open');
        if (animacjeWlaczone()) {
          requestAnimationFrame(() => appSettingsModal.classList.add('is-open'));
        } else {
          appSettingsModal.classList.add('is-open');
        }
      }
    }

    function zamknijAppSettings() {
      if (!appSettingsModal || appSettingsModal.hidden) return;
      if (
        appSettingsMobileMql.matches &&
        appSettingsModal.classList.contains('is-open') &&
        animacjeWlaczone()
      ) {
        appSettingsModal.classList.remove('is-open');
        const card = appSettingsModal.querySelector('.app-settings-card');
        if (!card) {
          ukryjAppSettingsModal();
          return;
        }
        let zamkniete = false;
        const finalize = () => {
          if (zamkniete) return;
          zamkniete = true;
          ukryjAppSettingsModal();
        };
        const done = (e) => {
          if (e.target !== card || e.propertyName !== 'transform') return;
          card.removeEventListener('transitionend', done);
          finalize();
        };
        card.addEventListener('transitionend', done);
        window.setTimeout(finalize, 320);
        return;
      }
      ukryjAppSettingsModal();
    }

    if (btnSettings) btnSettings.addEventListener('click', () => otworzAppSettings());
    if (btnSettingsMobile) {
      btnSettingsMobile.addEventListener('click', () => otworzAppSettings());
    }
    if (btnAppSettingsZamknij) btnAppSettingsZamknij.addEventListener('click', zamknijAppSettings);
    if (appSettingsBackdrop) appSettingsBackdrop.addEventListener('click', zamknijAppSettings);
    appSettingsMobileMql.addEventListener('change', () => {
      if (!appSettingsModal || appSettingsMobileMql.matches) return;
      appSettingsModal.classList.remove('is-open');
    });
    document.addEventListener('keydown', (e) => {
      if (e.key !== 'Escape') return;
      if (appSettingsModal && !appSettingsModal.hidden) zamknijAppSettings();
    });

    const appSettingsTheme = document.getElementById('app-settings-theme');
    if (appSettingsTheme) {
      appSettingsTheme.addEventListener('click', (e) => {
        const chip = e.target.closest('.chip[data-theme-pref]');
        if (!chip) return;
        ustawPreferencjeMotywu(chip.getAttribute('data-theme-pref'));
        odswiezChipyMotywuApp();
      });
    }

    const appSettingsValidity = document.getElementById('app-settings-validity');
    if (appSettingsValidity) {
      appSettingsValidity.addEventListener('click', (e) => {
        const chip = e.target.closest('.chip[data-validity-days]');
        if (!chip) return;
        const dni = Number(chip.getAttribute('data-validity-days'));
        zapiszDomyslnaWaznoscDni(dni);
        odswiezChipyWaznosciApp();
      });
    }

    const appSettingsDocType = document.getElementById('app-settings-doc-type');
    if (appSettingsDocType) {
      appSettingsDocType.addEventListener('click', (e) => {
        const chip = e.target.closest('.chip[data-default-doc-type]');
        if (!chip) return;
        const typ = chip.getAttribute('data-default-doc-type') || '';
        zapiszDomyslnyTypDokumentu(typ);
        odswiezChipyDocTypeApp();
      });
    }

    const appSettingsVat = document.getElementById('app-settings-vat');
    if (appSettingsVat) {
      appSettingsVat.addEventListener('click', (e) => {
        const chip = e.target.closest('.chip[data-default-vat]');
        if (!chip) return;
        zapiszDomyslnaStawkeVat(chip.getAttribute('data-default-vat'));
        odswiezChipyVatApp();
      });
    }

    migrujStaryConfig();

    const STORAGE_KEY_DRAFT = 'sumit_draft';
    const STORAGE_KEY_NUMERACJA = 'sumit_numeracja';
    const POLA_DRAFT = [
      'klient', 'numer_oferty', 'data_waznosci', 'uwagi',
      'numer_faktury', 'data_sprzedazy', 'termin_platnosci',
    ];
    const STORAGE_KEY_NUMERACJA_FAKTURY = 'sumit_numeracja_faktury';

    function wczytajNumeracje() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY_NUMERACJA);
        if (!raw) return { ostatniNumer: 0 };
        const dane = JSON.parse(raw);
        const n = Number(dane && dane.ostatniNumer);
        return { ostatniNumer: Number.isFinite(n) && n >= 0 ? Math.floor(n) : 0 };
      } catch (e) {
        return { ostatniNumer: 0 };
      }
    }

    function zapiszNumeracje(stan) {
      try {
        localStorage.setItem(STORAGE_KEY_NUMERACJA, JSON.stringify(stan));
      } catch (e) {}
    }

    function nastepnyNumerOferty() {
      const { ostatniNumer } = wczytajNumeracje();
      const rok = new Date().getFullYear();
      return `${rok}/${String(ostatniNumer + 1).padStart(3, '0')}`;
    }

    function inkrementujNumeracje() {
      const stan = wczytajNumeracje();
      stan.ostatniNumer = (Number(stan.ostatniNumer) || 0) + 1;
      zapiszNumeracje(stan);
    }

    function debounce(fn, wait) {
      let t;
      return function (...args) {
        clearTimeout(t);
        t = setTimeout(() => fn.apply(this, args), wait);
      };
    }

    function zbierzPozycjeDoDraft() {
      return [...tbody.querySelectorAll('tr')].map(tr => ({
        nazwa: tr.querySelector('.in-nazwa').value,
        ilosc: tr.querySelector('.in-ilosc').value,
        cena: tr.querySelector('.in-cena').value,
        koszt: tr.querySelector('.in-koszt').value,
        vat: tr.querySelector('.in-vat') ? tr.querySelector('.in-vat').value : '23',
      }));
    }

    function saveDraft() {
      const dane = { pozycje: zbierzPozycjeDoDraft() };
      POLA_DRAFT.forEach(id => {
        const el = document.getElementById(id);
        if (el) dane[id] = el.value;
      });
      dane._docType = getActiveDocType();
      try {
        localStorage.setItem(STORAGE_KEY_DRAFT, JSON.stringify(dane));
      } catch (e) {}
      aktualizujLivePodglad();
    }

    const saveDraftDebounced = debounce(saveDraft, 500);

    function wczytajDraft() {
      let dane;
      try {
        const raw = localStorage.getItem(STORAGE_KEY_DRAFT);
        if (!raw) return;
        dane = JSON.parse(raw);
      } catch (e) {
        return;
      }
      if (!dane || typeof dane !== 'object') return;

      POLA_DRAFT.forEach(id => {
        const el = document.getElementById(id);
        if (el && typeof dane[id] === 'string') el.value = dane[id];
      });

      if (Array.isArray(dane.pozycje) && dane.pozycje.length > 0) {
        tbody.innerHTML = '';
        dane.pozycje.forEach(p => {
          dodajWiersz({ naKoniec: true });
          const tr = tbody.lastElementChild;
          if (!tr) return;
          if (typeof p.nazwa === 'string') tr.querySelector('.in-nazwa').value = p.nazwa;
          if (typeof p.ilosc === 'string') tr.querySelector('.in-ilosc').value = p.ilosc;
          if (typeof p.cena === 'string')  tr.querySelector('.in-cena').value  = p.cena;
          if (typeof p.koszt === 'string') tr.querySelector('.in-koszt').value = p.koszt;
          if (typeof p.vat === 'string' && tr.querySelector('.in-vat')) tr.querySelector('.in-vat').value = p.vat;
        });
      }

      if (typeof dane._docType === 'string' && typeof setActiveDocType === 'function') {
        setActiveDocType(dane._docType);
      }

      odswiezStanPresetow();
      aktualizujSzacowanyZysk();
    }

    function ustawDomyslnaDate() {
      ustawDateZaDniOdDzis(wczytajDomyslnaWaznoscDni());
    }

    function wyczyscFormularz() {
      try { localStorage.removeItem(STORAGE_KEY_DRAFT); } catch (e) {}

      POLA_DRAFT.forEach(id => {
        const el = document.getElementById(id);
        if (el) el.value = '';
      });

      tbody.innerHTML = '';
      dodajWiersz();

      ustawDomyslnaDate();
      if (typeof setActiveDocType === 'function') setActiveDocType(wczytajDomyslnyTypDokumentu());

      const numerEl = document.getElementById('numer_oferty');
      if (numerEl) numerEl.value = nastepnyNumerOferty();

      odswiezStanPresetow();
      ukryjKomunikat();
      aktualizujSzacowanyZysk();
      aktualizujLivePodglad();
    }

    function oczyscNIPWejscie(s) {
      return String(s || '').replace(/\D/g, '');
    }

    const modalNIP = document.getElementById('modal-nip');
    const modalNIPBackdrop = document.getElementById('modal-nip-backdrop');
    const nipInput = document.getElementById('nip-input');
    const btnNipAnuluj = document.getElementById('btn-nip-anuluj');
    const btnNipPobierz = document.getElementById('btn-nip-pobierz');

    let nipFetchController = null;

    function ustawStanLadowaniaNIP(loading) {
      if (!btnNipPobierz || !nipInput) return;
      btnNipPobierz.disabled = loading;
      btnNipPobierz.textContent = loading ? 'Pobieram…' : 'Pobierz';
      btnNipPobierz.setAttribute('aria-busy', loading ? 'true' : 'false');
      nipInput.disabled = loading;
    }

    function otworzModalNIP() {
      if (!modalNIP || !nipInput) return;
      nipInput.value = '';
      ustawStanLadowaniaNIP(false);
      modalNIP.classList.remove('hidden');
      modalNIP.setAttribute('aria-hidden', 'false');
      document.body.style.overflow = 'hidden';
      setTimeout(() => nipInput.focus(), 0);
    }

    function zamknijModalNIP() {
      if (!modalNIP) return;
      if (nipFetchController) {
        nipFetchController.abort();
        nipFetchController = null;
      }
      modalNIP.classList.add('hidden');
      modalNIP.setAttribute('aria-hidden', 'true');
      document.body.style.overflow = '';
      ustawStanLadowaniaNIP(false);
    }

    async function wyslijZapytanieNIP() {
      const klientEl = document.getElementById('klient');
      if (!nipInput || !klientEl) return;

      const nip = oczyscNIPWejscie(nipInput.value);
      if (nip.length !== 10) {
        window.alert('Nieprawidłowy NIP — musi zawierać dokładnie 10 cyfr.');
        nipInput.focus();
        return;
      }

      ustawStanLadowaniaNIP(true);
      nipFetchController = new AbortController();

      try {
        const resp = await fetch('/api/nip?nip=' + encodeURIComponent(nip), {
          headers: { 'Accept': 'application/json' },
          signal: nipFetchController.signal,
        });
        let dane = null;
        try { dane = await resp.json(); } catch (e) { dane = null; }

        if (!resp.ok) {
          const msg = (dane && dane.error) ? dane.error : 'Nie udało się pobrać danych z Białej Listy MF.';
          window.alert(msg);
          return;
        }

        const linie = [];
        if (dane && dane.nazwa) linie.push(String(dane.nazwa));
        if (dane && dane.adres) linie.push(String(dane.adres));
        if (dane && dane.nip)   linie.push('NIP: ' + String(dane.nip));

        if (linie.length === 0) {
          window.alert('Biała Lista MF nie zwróciła danych firmy.');
          return;
        }

        klientEl.value = linie.join('\n');
        zamknijModalNIP();
        saveDraft();
        if (typeof zwinKlientaJesliWypelniony === 'function') zwinKlientaJesliWypelniony();
      } catch (err) {
        if (err && err.name === 'AbortError') return;
        window.alert('Błąd sieci: nie udało się połączyć z serwerem.');
      } finally {
        nipFetchController = null;
        if (!modalNIP.classList.contains('hidden')) {
          ustawStanLadowaniaNIP(false);
        }
      }
    }

    function pobierzKlientaPoNIP() {
      otworzModalNIP();
    }

    const klientTextarea = document.getElementById('klient');
    const klientAutocompleteEl = document.getElementById('klient-autocomplete');
    let autocompleteResults = [];
    let autocompleteIndex = -1;
    const MAX_PODPOWIEDZI_KLIENT = 8;

    function getUnikalniKlienci() {
      const seen = new Set();
      const wynik = [];
      const push = (blok) => {
        const klucz = String(blok || '').replace(/\s+/g, ' ').trim().toLowerCase();
        if (!klucz || seen.has(klucz)) return;
        seen.add(klucz);
        wynik.push(blok);
      };
      wczytajKlientow().forEach(k => push(formatKlientBlok(k)));
      wczytajHistorie().forEach(wpis => {
        const raw = (wpis && (wpis.klient || (wpis.payload && wpis.payload.klient))) || '';
        push(String(raw).trim());
      });
      return wynik;
    }

    function renderujAutocompleteKlient(matches) {
      klientAutocompleteEl.innerHTML = '';
      matches.forEach((klient, idx) => {
        const li = document.createElement('li');
        li.className = 'dropdown-item';
        li.setAttribute('role', 'option');
        li.dataset.idx = String(idx);
        const linie = klient.split('\n').map(s => s.trim()).filter(Boolean);
        const primary = document.createElement('span');
        primary.textContent = linie[0] || klient;
        li.appendChild(primary);
        if (linie.length > 1) {
          const meta = document.createElement('span');
          meta.className = 'dropdown-item-meta';
          meta.textContent = linie.slice(1).join(' · ');
          li.appendChild(meta);
        }
        li.addEventListener('mousedown', (e) => {
          e.preventDefault();
          wybierzKlientaZAutocomplete(klient);
        });
        klientAutocompleteEl.appendChild(li);
      });
    }

    function pokazAutocompleteKlient() {
      klientAutocompleteEl.classList.remove('hidden');
      if (klientTextarea) klientTextarea.setAttribute('aria-expanded', 'true');
    }

    function ukryjAutocompleteKlient() {
      klientAutocompleteEl.classList.add('hidden');
      autocompleteIndex = -1;
      if (klientTextarea) klientTextarea.setAttribute('aria-expanded', 'false');
    }

    function odswiezAutocompleteKlient() {
      if (!klientTextarea || !klientAutocompleteEl) return;
      const zapytanie = klientTextarea.value.trim().toLowerCase();
      if (zapytanie.length < 2) {
        ukryjAutocompleteKlient();
        return;
      }
      const unikalni = getUnikalniKlienci();
      const dopasowane = unikalni.filter(k => {
        const lower = k.toLowerCase();
        return lower.includes(zapytanie) && lower !== zapytanie;
      });
      if (dopasowane.length === 0) {
        ukryjAutocompleteKlient();
        return;
      }
      autocompleteResults = dopasowane.slice(0, MAX_PODPOWIEDZI_KLIENT);
      autocompleteIndex = -1;
      renderujAutocompleteKlient(autocompleteResults);
      pokazAutocompleteKlient();
    }

    function wybierzKlientaZAutocomplete(klient) {
      if (!klientTextarea || !klient) return;
      klientTextarea.dataset.klientAutofill = '1';
      klientTextarea.value = klient;
      delete klientTextarea.dataset.klientAutofill;
      ukryjAutocompleteKlient();
      saveDraft();
      if (typeof zwinKlientaJesliWypelniony === 'function') zwinKlientaJesliWypelniony();
      klientTextarea.blur();
    }

    function sprobujAutofillKlientaZKsiazki() {
      if (!klientTextarea) return false;
      if (klientTextarea.dataset.klientAutofill === '1') return false;
      const raw = klientTextarea.value;
      const linie = raw.split('\n').map(s => s.trim()).filter(Boolean);
      if (linie.length !== 1) return false;
      const trafiony = znajdzKlientaPoZapytaniu(linie[0]);
      if (!trafiony) return false;
      const blok = formatKlientBlok(trafiony);
      if (!blok || blok === raw) return false;
      klientTextarea.dataset.klientAutofill = '1';
      klientTextarea.value = blok;
      delete klientTextarea.dataset.klientAutofill;
      saveDraft();
      return true;
    }

    function obsluzInputKlienta() {
      if (klientTextarea && klientTextarea.dataset.klientAutofill === '1') return;
      if (sprobujAutofillKlientaZKsiazki()) {
        ukryjAutocompleteKlient();
        return;
      }
      odswiezAutocompleteKlient();
    }

    function ustawAktywnyItemAutocomplete(idx) {
      const items = klientAutocompleteEl.querySelectorAll('.dropdown-item');
      items.forEach((el, i) => {
        if (i === idx) {
          el.classList.add('is-active');
          el.scrollIntoView({ block: 'nearest' });
        } else {
          el.classList.remove('is-active');
        }
      });
      autocompleteIndex = idx;
    }

    if (klientTextarea && klientAutocompleteEl) {
      klientTextarea.addEventListener('input', obsluzInputKlienta);
      klientTextarea.addEventListener('focus', odswiezAutocompleteKlient);
      klientTextarea.addEventListener('keydown', (e) => {
        if (klientAutocompleteEl.classList.contains('hidden')) return;
        const items = klientAutocompleteEl.querySelectorAll('.dropdown-item');
        if (items.length === 0) return;
        if (e.key === 'ArrowDown') {
          e.preventDefault();
          const next = (autocompleteIndex + 1) % items.length;
          ustawAktywnyItemAutocomplete(next);
        } else if (e.key === 'ArrowUp') {
          e.preventDefault();
          const prev = autocompleteIndex <= 0 ? items.length - 1 : autocompleteIndex - 1;
          ustawAktywnyItemAutocomplete(prev);
        } else if (e.key === 'Enter') {
          if (autocompleteIndex >= 0 && autocompleteIndex < autocompleteResults.length) {
            e.preventDefault();
            wybierzKlientaZAutocomplete(autocompleteResults[autocompleteIndex]);
          }
        } else if (e.key === 'Escape') {
          ukryjAutocompleteKlient();
        }
      });
      document.addEventListener('click', (e) => {
        if (klientAutocompleteEl.classList.contains('hidden')) return;
        if (e.target === klientTextarea) return;
        if (klientAutocompleteEl.contains(e.target)) return;
        ukryjAutocompleteKlient();
      });
    }

    if (nipInput) {
      nipInput.addEventListener('input', () => {
        const cleaned = nipInput.value.replace(/\D/g, '');
        if (cleaned !== nipInput.value) nipInput.value = cleaned;
      });
      nipInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
          e.preventDefault();
          if (!btnNipPobierz.disabled) wyslijZapytanieNIP();
        }
      });
    }
    if (btnNipPobierz) btnNipPobierz.addEventListener('click', wyslijZapytanieNIP);
    if (btnNipAnuluj) btnNipAnuluj.addEventListener('click', zamknijModalNIP);
    if (modalNIPBackdrop) modalNIPBackdrop.addEventListener('click', zamknijModalNIP);
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && modalNIP && !modalNIP.classList.contains('hidden')) {
        zamknijModalNIP();
      }
    });

    const accordionMql = window.matchMedia('(max-width: 1023px)');
    const motionReduceMql = window.matchMedia('(prefers-reduced-motion: reduce)');

    function animacjeWlaczone() {
      return !motionReduceMql.matches;
    }

    function syncAccordionBodyHeight(section) {
      const body = section.querySelector('.accordion-body');
      if (!body) return;
      if (!accordionMql.matches) {
        body.style.maxHeight = '';
        return;
      }
      body.style.maxHeight = section.classList.contains('is-open') ? 'none' : '0px';
    }

    function ustawAkordeonOtwarty(section, otwarty) {
      const body = section.querySelector('.accordion-body');
      const header = section.querySelector('.accordion-header');
      if (!body) return;

      if (!accordionMql.matches) {
        body.style.maxHeight = '';
        section.classList.add('is-open');
        if (header) header.setAttribute('aria-expanded', 'true');
        return;
      }

      const animuj = animacjeWlaczone() && !accordionMql.matches;

      if (otwarty) {
        if (section.classList.contains('is-open') && body.style.maxHeight === 'none') return;

        section.classList.add('is-open');
        if (header) header.setAttribute('aria-expanded', 'true');
        const preview = section.querySelector('.accordion-preview');
        if (preview) preview.hidden = true;

        if (!animuj) {
          body.style.maxHeight = 'none';
          return;
        }

        body.style.maxHeight = '0px';
        requestAnimationFrame(() => {
          body.style.maxHeight = body.scrollHeight + 'px';
          const done = (e) => {
            if (e.propertyName !== 'max-height') return;
            body.removeEventListener('transitionend', done);
            if (section.classList.contains('is-open')) body.style.maxHeight = 'none';
          };
          body.addEventListener('transitionend', done);
        });
        return;
      }

      if (!section.classList.contains('is-open')) return;

      const odswiezPodglad = () => {
        if (section.dataset.accordionId === 'klient') {
          odswiezPodgladKlientaAkordeon();
        } else if (section.dataset.accordionId === 'pozycje') {
          odswiezPodgladPozycjiAkordeon();
        }
      };

      const zamknijPoAnimacji = () => {
        section.classList.remove('is-open');
        if (header) header.setAttribute('aria-expanded', 'false');
        body.style.maxHeight = '0px';
        odswiezPodglad();
      };

      if (!animuj) {
        body.style.maxHeight = '0px';
        zamknijPoAnimacji();
        return;
      }

      const h = body.scrollHeight;
      body.style.maxHeight = h + 'px';
      void body.offsetHeight;
      body.style.maxHeight = '0px';
      const done = (e) => {
        if (e.propertyName !== 'max-height') return;
        body.removeEventListener('transitionend', done);
        zamknijPoAnimacji();
      };
      body.addEventListener('transitionend', done);
    }

    function przelaczAkordeon(section) {
      if (!accordionMql.matches) return;
      ustawAkordeonOtwarty(section, !section.classList.contains('is-open'));
    }

    function odswiezPodgladKlientaAkordeon() {
      const section = document.querySelector('[data-accordion-id="klient"]');
      const klientEl = document.getElementById('klient');
      if (!section || !klientEl) return;
      const preview = section.querySelector('.accordion-preview');
      if (!preview) return;

      if (!accordionMql.matches || section.classList.contains('is-open')) {
        preview.hidden = true;
        preview.textContent = '';
        return;
      }

      const pierwszaLinia = klientEl.value.split('\n').map(s => s.trim()).find(Boolean) || '';
      if (pierwszaLinia) {
        preview.textContent = '\u2713 ' + pierwszaLinia;
        preview.hidden = false;
      } else {
        preview.textContent = '';
        preview.hidden = true;
      }
    }

    function odswiezPodgladPozycjiAkordeon() {
      if (!accordionFormReady) return;
      const section = document.querySelector('[data-accordion-id="pozycje"]');
      if (!section) return;
      const preview = section.querySelector('.accordion-preview');
      if (!preview) return;

      if (!accordionMql.matches || section.classList.contains('is-open')) {
        preview.hidden = true;
        preview.textContent = '';
        return;
      }

      const nazwy = [...tbody.querySelectorAll('tr')]
        .map(tr => (tr.querySelector('.in-nazwa')?.value || '').trim())
        .filter(Boolean);

      if (nazwy.length === 0) {
        preview.textContent = '';
        preview.hidden = true;
        return;
      }

      if (nazwy.length === 1) {
        preview.textContent = '\u2713 ' + nazwy[0];
      } else {
        const n = nazwy.length;
        const mod10 = n % 10;
        const mod100 = n % 100;
        const odmiana = (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) ? 'pozycje' : 'pozycji';
        preview.textContent = '\u2713 ' + n + ' ' + odmiana;
      }
      preview.hidden = false;
    }

    function zwinKlientaJesliWypelniony() {
      if (!accordionMql.matches) return;
      const section = document.querySelector('[data-accordion-id="klient"]');
      const klientEl = document.getElementById('klient');
      if (!section || !klientEl) return;
      if (klientEl.value.trim() && section.classList.contains('is-open')) {
        ustawAkordeonOtwarty(section, false);
      } else {
        odswiezPodgladKlientaAkordeon();
      }
    }

    function odswiezWszystkieAkordeony() {
      document.querySelectorAll('#oferta-form .accordion-section').forEach(section => {
        const body = section.querySelector('.accordion-body');
        const header = section.querySelector('.accordion-header');
        if (!body) return;
        if (!accordionMql.matches) {
          body.style.maxHeight = '';
          section.classList.add('is-open');
          if (header) header.setAttribute('aria-expanded', 'true');
        } else {
          syncAccordionBodyHeight(section);
        }
      });
      odswiezPodgladKlientaAkordeon();
      odswiezPodgladPozycjiAkordeon();
    }

    function initFormAccordions() {
      const sections = document.querySelectorAll('#oferta-form .accordion-section');
      const klientEl = document.getElementById('klient');

      sections.forEach(section => {
        const domyslnieOtwarty = section.dataset.accordionDefault === 'open';
        const body = section.querySelector('.accordion-body');
        const header = section.querySelector('.accordion-header');

        section.classList.toggle('is-open', domyslnieOtwarty);
        if (header) header.setAttribute('aria-expanded', domyslnieOtwarty ? 'true' : 'false');

        if (body) {
          if (!accordionMql.matches) {
            section.classList.add('is-open');
            body.style.maxHeight = '';
          } else {
            syncAccordionBodyHeight(section);
          }
        }

        if (header) {
          header.addEventListener('click', () => przelaczAkordeon(section));
        }
      });

      if (klientEl) {
        klientEl.addEventListener('blur', zwinKlientaJesliWypelniony);
      }

      accordionMql.addEventListener('change', odswiezWszystkieAkordeony);

      if (klientEl && klientEl.value.trim()) {
        const klientSection = document.querySelector('[data-accordion-id="klient"]');
        if (klientSection && accordionMql.matches) {
          ustawAkordeonOtwarty(klientSection, false);
        }
      }
      accordionFormReady = true;
      odswiezPodgladKlientaAkordeon();
      odswiezPodgladPozycjiAkordeon();
    }

    function initMobileStickyHeader() {
      const chrome = document.getElementById('mobile-chrome');
      const mql = window.matchMedia('(max-width: 1023px)');
      if (!chrome) return;

      let ticking = false;
      function odswiez() {
        if (!mql.matches) {
          chrome.classList.remove('is-scrolled');
          return;
        }
        const y = window.scrollY || document.documentElement.scrollTop || 0;
        chrome.classList.toggle('is-scrolled', y > 6);
      }

      function onScroll() {
        if (ticking) return;
        ticking = true;
        requestAnimationFrame(() => {
          odswiez();
          ticking = false;
        });
      }

      window.addEventListener('scroll', onScroll, { passive: true });
      mql.addEventListener('change', odswiez);
      odswiez();
    }

    function initMobileAiSheet() {
      const fab = document.getElementById('btn-ai-fab');
      const modal = document.getElementById('ai-input-modal');
      const backdrop = document.getElementById('ai-input-modal-backdrop');
      const closeBtn = document.getElementById('btn-ai-sheet-close');
      const notatka = document.getElementById('ai-notatka');
      const mql = window.matchMedia('(max-width: 1023px)');
      if (!modal) return;

      function zamknij() {
        if (!mql.matches) return;
        modal.classList.remove('is-open');
        modal.classList.add('hidden');
        modal.setAttribute('hidden', '');
        modal.setAttribute('aria-hidden', 'true');
        if (!document.querySelector('.modal:not(.hidden):not([hidden])')) {
          document.body.style.overflow = '';
        }
      }

      function otworz() {
        if (!mql.matches) return;
        modal.classList.remove('hidden');
        modal.removeAttribute('hidden');
        modal.setAttribute('aria-hidden', 'false');
        requestAnimationFrame(() => {
          modal.classList.add('is-open');
          notatka?.focus();
        });
        document.body.style.overflow = 'hidden';
      }

      function syncAiSheetLayout() {
        if (!mql.matches) {
          modal.classList.remove('is-open', 'hidden');
          modal.removeAttribute('hidden');
          modal.setAttribute('aria-hidden', 'false');
          const rozwiniety = !modal.classList.contains('is-collapsed');
          fab?.setAttribute('aria-expanded', rozwiniety ? 'true' : 'false');
          return;
        }
        if (!modal.classList.contains('is-open')) {
          modal.classList.add('hidden');
          modal.setAttribute('hidden', '');
          modal.setAttribute('aria-hidden', 'true');
        }
        modal.classList.remove('is-collapsed');
        fab?.removeAttribute('aria-expanded');
      }

      fab?.addEventListener('click', () => {
        if (mql.matches) {
          otworz();
          return;
        }
        modal.classList.toggle('is-collapsed');
        const rozwiniety = !modal.classList.contains('is-collapsed');
        fab?.setAttribute('aria-expanded', rozwiniety ? 'true' : 'false');
        if (rozwiniety) notatka?.focus();
      });
      backdrop?.addEventListener('click', zamknij);
      closeBtn?.addEventListener('click', zamknij);

      document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && modal.classList.contains('is-open')) {
          zamknij();
        }
      });

      mql.addEventListener('change', () => {
        syncAiSheetLayout();
      });

      syncAiSheetLayout();

      window.zamknijAiInputSheet = zamknij;
    }

    function initMobileScrollClamp() {
      const mql = window.matchMedia('(max-width: 1023px)');

      function clamp() {
        if (!mql.matches) return;
        const view = document.body.dataset.activeView;
        let root = null;
        if (view === 'view-kreator') root = document.getElementById('oferta-form');
        else if (view === 'view-firma') root = document.getElementById('view-firma');
        else if (view === 'view-statystyki') root = document.getElementById('view-statystyki');
        if (!root || root.classList.contains('hidden') || root.hasAttribute('hidden')) return;
        const maxScroll = Math.max(0, root.offsetTop + root.offsetHeight - window.innerHeight);
        if (window.scrollY > maxScroll + 1) window.scrollTo(0, maxScroll);
      }

      window.addEventListener('scroll', clamp, { passive: true });
      window.addEventListener('resize', clamp, { passive: true });
      mql.addEventListener('change', clamp);
    }

    document.addEventListener('DOMContentLoaded', () => {
      // Viewer mode: ?w= → renderuj read-only podgląd dla klienta, pomiń init edytora
      const _wParam = new URLSearchParams(location.search).get('w');
      if (_wParam) {
        trackEvent('client_view');
        initKlientViewer(_wParam);
        return;
      }

      trackEvent('page_view');

      wczytajDraft();

      try {
        if (!localStorage.getItem(STORAGE_KEY_DRAFT) && typeof setActiveDocType === 'function') {
          setActiveDocType(wczytajDomyslnyTypDokumentu());
        }
      } catch (e) {}

      const numerEl = document.getElementById('numer_oferty');
      if (numerEl && !numerEl.value.trim()) {
        numerEl.value = nastepnyNumerOferty();
      }

      initFormAccordions();
      initAiQuickInput();
      initMobileStickyHeader();
      initMobileAiSheet();
      initMobileScrollClamp();

      form.addEventListener('input', saveDraftDebounced);
      btnDodaj.addEventListener('click', saveDraftDebounced);
      tbody.addEventListener('click', (e) => {
        if (e.target && e.target.classList && e.target.classList.contains('btn-remove')) {
          saveDraftDebounced();
        }
      });

      document.getElementById('btn-wyczysc-formularz').addEventListener('click', wyczyscFormularz);

      const btnPobierzNIP = document.getElementById('btn-pobierz-nip');
      if (btnPobierzNIP) btnPobierzNIP.addEventListener('click', pobierzKlientaPoNIP);

      renderStatystyki();
      initViewTabs();
      aktualizujSzacowanyZysk();
      aktualizujLivePodglad();
      synchronizujWysokoscPodgladu();

      const btnKopiujLink = document.getElementById('btn-kopiuj-link');
      if (btnKopiujLink) btnKopiujLink.addEventListener('click', skopiujLinkWyceny);

      const btnPobierzXml = document.getElementById('btn-pobierz-xml');
      if (btnPobierzXml) btnPobierzXml.addEventListener('click', pobierzXmlKSeF);

      initDocTypeSwitcher();
    });

    function initViewTabs() {
      const tabs = document.querySelectorAll('.view-tab[data-view-target]');
      if (!tabs.length) return;

      const widoki = new Map();
      tabs.forEach(tab => {
        const targetId = tab.getAttribute('data-view-target');
        const widok = document.getElementById(targetId);
        if (widok) widoki.set(targetId, widok);
      });

      function aktywuj(targetId) {
        if (widokFirmaAktywny() && targetId !== 'view-firma') {
          if (!zapiszFirmeJesliTrzeba()) return;
        }
        const poprzedniWidok = document.body.dataset.activeView || 'view-kreator';
        document.body.dataset.activeView = targetId;
        tabs.forEach(tab => {
          const aktywny = tab.getAttribute('data-view-target') === targetId;
          tab.classList.toggle('is-active', aktywny);
          tab.setAttribute('aria-selected', aktywny ? 'true' : 'false');
        });

        let widokWejscia = null;
        widoki.forEach((widok, id) => {
          const aktywny = id === targetId;
          if (aktywny) {
            widokWejscia = widok;
            widok.classList.remove('hidden');
            widok.removeAttribute('hidden');
          } else {
            widok.classList.add('hidden');
            widok.setAttribute('hidden', '');
            widok.classList.remove('view-enter');
          }
        });

        const viewTabsMobileMql = window.matchMedia('(max-width: 1023px)');
        if (widokWejscia && animacjeWlaczone() && !viewTabsMobileMql.matches) {
          widokWejscia.classList.remove('view-enter');
          void widokWejscia.offsetWidth;
          widokWejscia.classList.add('view-enter');
          widokWejscia.addEventListener('animationend', function onViewEnterEnd() {
            widokWejscia.classList.remove('view-enter');
            widokWejscia.removeEventListener('animationend', onViewEnterEnd);
          });
        } else if (widokWejscia) {
          widokWejscia.classList.remove('view-enter');
        }

        if (widokWejscia) {
          window.scrollTo({
            top: 0,
            behavior: viewTabsMobileMql.matches ? 'auto' : (animacjeWlaczone() ? 'smooth' : 'auto'),
          });
        }

        if (targetId === 'view-statystyki') {
          renderStatystyki();
        }
        if (targetId === 'view-firma' && poprzedniWidok !== 'view-firma') {
          wypelnijFormularzCfg();
        }
      }

      aktywujWidok = aktywuj;
      document.body.dataset.activeView = 'view-kreator';

      tabs.forEach(tab => {
        tab.addEventListener('click', () => {
          const targetId = tab.getAttribute('data-view-target');
          if (targetId) aktywuj(targetId);
        });
      });
    }

    const STORAGE_KEY_KATALOG = 'sumit_katalog';
    const STORAGE_KEY_KATALOG_VER = 'sumit_katalog_v';
    const KATALOG_VERSION = 3;

    const DOMYSLNY_KATALOG = [
      { nazwa: 'Wymiana baterii umywalkowej',       cena_jednostkowa: 150, kategoria: 'Hydraulika' },
      { nazwa: 'Naprawa spłuczki',                  cena_jednostkowa: 100, kategoria: 'Hydraulika' },
      { nazwa: 'Udrożnienie odpływu',               cena_jednostkowa: 200, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż umywalki',                   cena_jednostkowa: 250, kategoria: 'Hydraulika' },
      { nazwa: 'Wymiana zaworu kątowego',           cena_jednostkowa:  80, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż WC kompakt',                 cena_jednostkowa: 350, kategoria: 'Hydraulika' },
      { nazwa: 'Wymiana grzejnika',                 cena_jednostkowa: 450, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż baterii prysznicowej',       cena_jednostkowa: 180, kategoria: 'Hydraulika' },
      { nazwa: 'Wymiana rury',                      cena_jednostkowa: 150, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż pompy cyrkulacyjnej',        cena_jednostkowa: 400, kategoria: 'Hydraulika' },
      { nazwa: 'Uszczelnienie pionu',               cena_jednostkowa: 320, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż filtra wody',                cena_jednostkowa: 180, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż podgrzewacza wody',         cena_jednostkowa: 380, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż wanny',                      cena_jednostkowa: 480, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż brodzika prysznicowego',     cena_jednostkowa: 320, kategoria: 'Hydraulika' },
      { nazwa: 'Podłączenie zmywarki',              cena_jednostkowa: 120, kategoria: 'Hydraulika' },
      { nazwa: 'Podłączenie pralki',                cena_jednostkowa: 100, kategoria: 'Hydraulika' },
      { nazwa: 'Wymiana syfonu',                    cena_jednostkowa:  90, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż bidetu',                     cena_jednostkowa: 280, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż odpływu liniowego',          cena_jednostkowa: 240, kategoria: 'Hydraulika' },
      { nazwa: 'Instalacja prysznica bezbrodzikowego', cena_jednostkowa: 580, kategoria: 'Hydraulika' },
      { nazwa: 'Wymiana zaworu trójdrożnego',       cena_jednostkowa:  70, kategoria: 'Hydraulika' },
      { nazwa: 'Czyszczenie kanalizacji',           cena_jednostkowa: 220, kategoria: 'Hydraulika' },
      { nazwa: 'Montaż stelaża podtynkowego WC',    cena_jednostkowa: 420, kategoria: 'Hydraulika' },

      { nazwa: 'Montaż gniazda elektrycznego',      cena_jednostkowa:  80, kategoria: 'Elektryka' },
      { nazwa: 'Wymiana włącznika',                 cena_jednostkowa:  60, kategoria: 'Elektryka' },
      { nazwa: 'Montaż lampy sufitowej',            cena_jednostkowa: 120, kategoria: 'Elektryka' },
      { nazwa: 'Montaż żyrandola',                  cena_jednostkowa: 150, kategoria: 'Elektryka' },
      { nazwa: 'Pomiary elektryczne (protokół)',    cena_jednostkowa: 200, kategoria: 'Elektryka' },
      { nazwa: 'Wymiana rozdzielni',                cena_jednostkowa: 800, kategoria: 'Elektryka' },
      { nazwa: 'Montaż rozdzielni',                 cena_jednostkowa: 1200, kategoria: 'Elektryka' },
      { nazwa: 'Wymiana bezpieczników',             cena_jednostkowa:  80, kategoria: 'Elektryka' },
      { nazwa: 'Montaż oświetlenia LED',            cena_jednostkowa: 100, kategoria: 'Elektryka' },
      { nazwa: 'Montaż wentylatora łazienkowego',   cena_jednostkowa: 180, kategoria: 'Elektryka' },
      { nazwa: 'Pomiar rezystancji izolacji',       cena_jednostkowa: 250, kategoria: 'Elektryka' },
      { nazwa: 'Montaż alarmu',                     cena_jednostkowa: 650, kategoria: 'Elektryka' },
      { nazwa: 'Montaż domofonu',                   cena_jednostkowa: 350, kategoria: 'Elektryka' },
      { nazwa: 'Montaż klimatyzacji (podłączenie)', cena_jednostkowa: 480, kategoria: 'Elektryka' },
      { nazwa: 'Montaż wideodomofonu',              cena_jednostkowa: 450, kategoria: 'Elektryka' },
      { nazwa: 'Prowadzenie przewodu (mb)',         cena_jednostkowa:  35, kategoria: 'Elektryka' },
      { nazwa: 'Montaż termostatu',                 cena_jednostkowa: 120, kategoria: 'Elektryka' },
      { nazwa: 'Instalacja czujnika ruchu',         cena_jednostkowa:  70, kategoria: 'Elektryka' },
      { nazwa: 'Wymiana bezpiecznika różnicowoprądowego', cena_jednostkowa: 150, kategoria: 'Elektryka' },
      { nazwa: 'Podłączenie płyty indukcyjnej',     cena_jednostkowa: 200, kategoria: 'Elektryka' },
      { nazwa: 'Montaż listwy LED (mb)',             cena_jednostkowa:  80, kategoria: 'Elektryka' },
      { nazwa: 'Przegląd instalacji elektrycznej',  cena_jednostkowa: 300, kategoria: 'Elektryka' },
      { nazwa: 'Montaż gniazda z uziemieniem',      cena_jednostkowa:  90, kategoria: 'Elektryka' },
      { nazwa: 'Podłączenie bojlera elektrycznego', cena_jednostkowa: 220, kategoria: 'Elektryka' },

      { nazwa: 'Malowanie ściany (m²)',             cena_jednostkowa:  35, kategoria: 'Wykończenia' },
      { nazwa: 'Gładź gipsowa (m²)',                cena_jednostkowa:  50, kategoria: 'Wykończenia' },
      { nazwa: 'Tapetowanie (m²)',                  cena_jednostkowa:  45, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż drzwi wewnętrznych',         cena_jednostkowa: 350, kategoria: 'Wykończenia' },
      { nazwa: 'Listwy przypodłogowe (mb)',         cena_jednostkowa:  25, kategoria: 'Wykończenia' },
      { nazwa: 'Układanie paneli (m²)',             cena_jednostkowa:  40, kategoria: 'Wykończenia' },
      { nazwa: 'Układanie płytek (m²)',             cena_jednostkowa: 120, kategoria: 'Wykończenia' },
      { nazwa: 'Fugowanie (m²)',                    cena_jednostkowa:  35, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż parapetu',                   cena_jednostkowa:  80, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż sufitu podwieszanego (m²)',  cena_jednostkowa:  95, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż listwy sufitowej (mb)',     cena_jednostkowa:  20, kategoria: 'Wykończenia' },
      { nazwa: 'Szpachlowanie nierówności (m²)',    cena_jednostkowa:  30, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż progów drzwiowych',          cena_jednostkowa:  60, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż rolet wewnętrznych',         cena_jednostkowa:  90, kategoria: 'Wykończenia' },
      { nazwa: 'Układanie paneli ściennych (m²)',   cena_jednostkowa:  55, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż ościeżnic',                  cena_jednostkowa: 120, kategoria: 'Wykończenia' },
      { nazwa: 'Cyklinowanie podłogi (m²)',         cena_jednostkowa:  45, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż zabudowy GK (m²)',           cena_jednostkowa:  85, kategoria: 'Wykończenia' },
      { nazwa: 'Lakierowanie podłogi (m²)',         cena_jednostkowa:  50, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż lamperii (m²)',              cena_jednostkowa:  65, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż karnisza',                   cena_jednostkowa:  70, kategoria: 'Wykończenia' },
      { nazwa: 'Montaż listew maskujących (mb)',    cena_jednostkowa:  18, kategoria: 'Wykończenia' },
      { nazwa: 'Układanie mozaiki (m²)',            cena_jednostkowa: 150, kategoria: 'Wykończenia' },

      { nazwa: 'Wyburzenie ścianki działowej (m²)', cena_jednostkowa: 180, kategoria: 'Budowlanka' },
      { nazwa: 'Stawianie ścianki działowej (m²)',  cena_jednostkowa: 220, kategoria: 'Budowlanka' },
      { nazwa: 'Murowanie ściany (m²)',             cena_jednostkowa: 180, kategoria: 'Budowlanka' },
      { nazwa: 'Wylewka betonowa (m²)',             cena_jednostkowa:  55, kategoria: 'Budowlanka' },
      { nazwa: 'Montaż rusztowania (dzień)',        cena_jednostkowa: 350, kategoria: 'Budowlanka' },
      { nazwa: 'Tynkowanie (m²)',                   cena_jednostkowa:  45, kategoria: 'Budowlanka' },
      { nazwa: 'Izolacja fundamentów (mb)',         cena_jednostkowa: 120, kategoria: 'Budowlanka' },
      { nazwa: 'Wylewka anhydrytowa (m²)',          cena_jednostkowa:  65, kategoria: 'Budowlanka' },
      { nazwa: 'Wylewka samopoziomująca (m²)',      cena_jednostkowa:  40, kategoria: 'Budowlanka' },
      { nazwa: 'Ocieplenie styropianem (m²)',       cena_jednostkowa:  75, kategoria: 'Budowlanka' },
      { nazwa: 'Izolacja termiczna budynku (m²)',   cena_jednostkowa:  90, kategoria: 'Budowlanka' },
      { nazwa: 'Wykuwanie otworu w ścianie',        cena_jednostkowa: 380, kategoria: 'Budowlanka' },
      { nazwa: 'Montaż okna budowlanego',           cena_jednostkowa: 280, kategoria: 'Budowlanka' },
      { nazwa: 'Betonowanie schodów (stopień)',     cena_jednostkowa: 180, kategoria: 'Budowlanka' },
      { nazwa: 'Montaż belki stropowej',            cena_jednostkowa: 450, kategoria: 'Budowlanka' },
      { nazwa: 'Wzmacnianie stropu',                cena_jednostkowa: 850, kategoria: 'Budowlanka' },
      { nazwa: 'Płyta fundamentowa (m²)',           cena_jednostkowa: 220, kategoria: 'Budowlanka' },
      { nazwa: 'Murowanie komina',                  cena_jednostkowa: 2800, kategoria: 'Budowlanka' },
      { nazwa: 'Montaż dźwigaru stalowego',         cena_jednostkowa: 650, kategoria: 'Budowlanka' },

      { nazwa: 'Robocizna (godzina)',               cena_jednostkowa:  80, kategoria: 'Inne' },
      { nazwa: 'Dojazd (km)',                       cena_jednostkowa:   3, kategoria: 'Inne' },
      { nazwa: 'Wywóz gruzu (m³)',                  cena_jednostkowa: 350, kategoria: 'Inne' },
      { nazwa: 'Transport materiałów',              cena_jednostkowa: 150, kategoria: 'Inne' },
      { nazwa: 'Dojazd do klienta (ryczałt)',       cena_jednostkowa:  50, kategoria: 'Inne' },
      { nazwa: 'Wycena na miejscu',                 cena_jednostkowa: 100, kategoria: 'Inne' },
      { nazwa: 'Konsultacja techniczna (godz.)',    cena_jednostkowa: 120, kategoria: 'Inne' },
      { nazwa: 'Przygotowanie kosztorysu',          cena_jednostkowa: 250, kategoria: 'Inne' },
      { nazwa: 'Dokumentacja powykonawcza',         cena_jednostkowa: 200, kategoria: 'Inne' },
      { nazwa: 'Utylizacja odpadów budowlanych (m³)', cena_jednostkowa: 320, kategoria: 'Inne' },
      { nazwa: 'Prace w weekend / święta (dopłata)', cena_jednostkowa: 150, kategoria: 'Inne' },
      { nazwa: 'Opłata za pilne zlecenie',          cena_jednostkowa: 100, kategoria: 'Inne' },
      { nazwa: 'Demontaż starych elementów (godz.)', cena_jednostkowa:  90, kategoria: 'Inne' },
      { nazwa: 'Nadzór budowlany (godz.)',          cena_jednostkowa: 100, kategoria: 'Inne' },
      { nazwa: 'Wynajem narzędzi (dzień)',          cena_jednostkowa:  80, kategoria: 'Inne' },
      { nazwa: 'Magazynowanie materiałów (dzień)',  cena_jednostkowa:  60, kategoria: 'Inne' },
      { nazwa: 'Odbiór techniczny',                 cena_jednostkowa: 300, kategoria: 'Inne' },
      { nazwa: 'Pakiet materiałów pomocniczych',     cena_jednostkowa:  80, kategoria: 'Inne' },
    ];

    function zapiszKatalog(items) {
      try {
        localStorage.setItem(STORAGE_KEY_KATALOG, JSON.stringify(items));
        localStorage.setItem(STORAGE_KEY_KATALOG_VER, String(KATALOG_VERSION));
      } catch (e) {}
    }

    function wczytajKatalog() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY_KATALOG);
        const ver = parseInt(localStorage.getItem(STORAGE_KEY_KATALOG_VER) || '0', 10);
        if (!raw || ver < KATALOG_VERSION) {
          zapiszKatalog(DOMYSLNY_KATALOG);
          return DOMYSLNY_KATALOG.slice();
        }
        const dane = JSON.parse(raw);
        if (!Array.isArray(dane)) return DOMYSLNY_KATALOG.slice();
        return dane
          .filter(it => it && typeof it.nazwa === 'string')
          .map(it => ({
            nazwa: it.nazwa,
            cena_jednostkowa: Number(it.cena_jednostkowa) || 0,
            kategoria: (typeof it.kategoria === 'string' && it.kategoria.trim()) || 'Inne',
          }));
      } catch (e) {
        return DOMYSLNY_KATALOG.slice();
      }
    }

    const catalogModal = document.getElementById('catalog-modal');
    const catalogBackdrop = document.getElementById('catalog-modal-backdrop');
    const catalogList = document.getElementById('catalog-list');
    const catalogSearch = document.getElementById('catalog-search');
    const catalogCategories = document.getElementById('catalog-categories');
    const catalogFeedback = document.getElementById('catalog-feedback');
    const btnKatalog = document.getElementById('btn-katalog');
    const btnZamknijCatalog = document.getElementById('btn-zamknij-catalog');
    const btnImportCatalog = document.getElementById('btn-import-catalog');
    const btnExportCatalog = document.getElementById('btn-export-catalog');
    const importCatalogInput = document.getElementById('import-catalog');

    let aktywnyWiersz = null;
    let aktywnaKategoria = '';

    tbody.addEventListener('focusin', (e) => {
      const tr = e.target && e.target.closest && e.target.closest('tr');
      if (tr && tbody.contains(tr)) aktywnyWiersz = tr;
    });

    function formatujCene(cena) {
      const n = Number(cena);
      if (!Number.isFinite(n)) return '';
      return n.toFixed(2).replace('.', ',') + ' zł';
    }

    function pokazFeedbackKatalog(tekst, typ) {
      if (!catalogFeedback) return;
      catalogFeedback.textContent = tekst;
      catalogFeedback.className = 'catalog-feedback' + (typ === 'error' ? ' error' : '');
      catalogFeedback.hidden = false;
      if (typ !== 'error') {
        setTimeout(() => { if (catalogFeedback.textContent === tekst) ukryjFeedbackKatalog(); }, 3000);
      }
    }

    function ukryjFeedbackKatalog() {
      if (!catalogFeedback) return;
      catalogFeedback.hidden = true;
      catalogFeedback.textContent = '';
      catalogFeedback.className = 'catalog-feedback';
    }

    function dostepneKategorie(items) {
      const set = new Set();
      items.forEach(it => set.add(it.kategoria || 'Inne'));
      return Array.from(set);
    }

    function renderujKategorie(items) {
      catalogCategories.innerHTML = '';
      const kategorie = dostepneKategorie(items);
      const wszystkie = document.createElement('button');
      wszystkie.type = 'button';
      wszystkie.className = 'chip';
      wszystkie.textContent = 'Wszystkie';
      wszystkie.setAttribute('aria-pressed', aktywnaKategoria === '' ? 'true' : 'false');
      wszystkie.addEventListener('click', () => {
        aktywnaKategoria = '';
        renderujKatalog();
      });
      catalogCategories.appendChild(wszystkie);
      kategorie.forEach(kat => {
        const chip = document.createElement('button');
        chip.type = 'button';
        chip.className = 'chip';
        chip.textContent = kat;
        chip.setAttribute('aria-pressed', aktywnaKategoria === kat ? 'true' : 'false');
        chip.addEventListener('click', () => {
          aktywnaKategoria = aktywnaKategoria === kat ? '' : kat;
          renderujKatalog();
        });
        catalogCategories.appendChild(chip);
      });
    }

    function renderujKatalog() {
      const items = wczytajKatalog();
      renderujKategorie(items);

      const q = (catalogSearch.value || '').trim().toLowerCase();
      const widoczne = items.filter(it => {
        if (aktywnaKategoria && (it.kategoria || 'Inne') !== aktywnaKategoria) return false;
        if (q && !it.nazwa.toLowerCase().includes(q)) return false;
        return true;
      });

      catalogList.innerHTML = '';
      if (widoczne.length === 0) {
        const li = document.createElement('li');
        li.className = 'catalog-empty';
        li.textContent = 'Brak pozycji pasujących do filtra.';
        catalogList.appendChild(li);
        return;
      }
      widoczne.forEach(item => {
        const li = document.createElement('li');
        li.className = 'catalog-item';
        li.setAttribute('role', 'option');
        li.tabIndex = 0;
        li.dataset.kategoria = item.kategoria || 'Inne';
        const nazwa = document.createElement('span');
        nazwa.className = 'catalog-item-nazwa';
        nazwa.textContent = item.nazwa;
        const cena = document.createElement('span');
        cena.className = 'catalog-item-cena';
        cena.textContent = formatujCene(item.cena_jednostkowa);
        li.appendChild(nazwa);
        li.appendChild(cena);
        li.addEventListener('click', () => addFromCatalog(item));
        li.addEventListener('keydown', (e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            addFromCatalog(item);
          }
        });
        catalogList.appendChild(li);
      });
    }

    function addFromCatalog(item) {
      if (!item || typeof item !== 'object') return;
      dodajWiersz();
      const tr = tbody.firstElementChild;
      if (!tr) return;

      const inNazwa = tr.querySelector('.in-nazwa');
      const inIlosc = tr.querySelector('.in-ilosc');
      const inCena  = tr.querySelector('.in-cena');

      if (inNazwa && typeof item.nazwa === 'string') {
        inNazwa.value = item.nazwa;
        inNazwa.dispatchEvent(new Event('input', { bubbles: true }));
      }
      if (inIlosc && !inIlosc.value.trim()) {
        inIlosc.value = '1';
        inIlosc.dispatchEvent(new Event('input', { bubbles: true }));
      }
      if (inCena) {
        const cena = Number(item.cena_jednostkowa);
        if (Number.isFinite(cena)) {
          inCena.value = String(cena);
          inCena.dispatchEvent(new Event('input', { bubbles: true }));
        }
      }

      aktywnyWiersz = tr;
      zamknijKatalog();
    }

    window.addFromCatalog = addFromCatalog;

    const aiNotatka = document.getElementById('ai-notatka');
    const btnAiParse = document.getElementById('btn-ai-parse');
    const btnAiMic = document.getElementById('btn-ai-mic');
    const btnAiPhoto = document.getElementById('btn-ai-photo');
    const aiPhotoInput = document.getElementById('ai-photo-input');
    const aiPhotoThumb = document.getElementById('ai-photo-thumb');
    const aiPhotoThumbImg = document.getElementById('ai-photo-thumb-img');
    const btnAiPhotoRemove = document.getElementById('btn-ai-photo-remove');
    const aiParseStatus = document.getElementById('ai-parse-status');
    const btnAiParseLabel = btnAiParse ? btnAiParse.querySelector('.btn-ai-parse-label') : null;
    const btnAiParseSpinner = btnAiParse ? btnAiParse.querySelector('.btn-ai-parse-spinner') : null;
    const modalAiPreview = document.getElementById('modal-ai-preview');
    const modalAiPreviewBackdrop = document.getElementById('modal-ai-preview-backdrop');
    const aiPreviewTbody = document.getElementById('ai-preview-tbody');
    const aiPreviewMeta = document.getElementById('ai-preview-meta');
    const btnAiPreviewAnuluj = document.getElementById('btn-ai-preview-anuluj');
    const btnAiPreviewDodaj = document.getElementById('btn-ai-preview-dodaj');

    const AI_MAX_KONTEKST = 30;
    const AI_PHOTO_MAX_BYTES = 5 * 1024 * 1024;

    let aiParseController = null;
    let speechRecognition = null;
    let speechRecording = false;
    let speechBaseText = '';
    let aiPendingItems = [];
    let aiPhotoPayload = null;

    function ustawAiParseStatus(tekst, typ) {
      if (!aiParseStatus) return;
      aiParseStatus.textContent = tekst || '';
      aiParseStatus.className = 'ai-parse-status' + (typ ? ' is-' + typ : '');
    }

    function ustawStanLadowaniaAiParse(loading) {
      if (!btnAiParse) return;
      btnAiParse.disabled = loading;
      btnAiParse.setAttribute('aria-busy', loading ? 'true' : 'false');
      if (btnAiMic) btnAiMic.disabled = loading;
      if (btnAiPhoto) btnAiPhoto.disabled = loading;
      if (btnAiPhotoRemove) btnAiPhotoRemove.disabled = loading;
      if (btnAiParseLabel) {
        btnAiParseLabel.textContent = loading ? 'Przygotowuję…' : 'Pokaż propozycję';
      }
      if (btnAiParseSpinner) {
        btnAiParseSpinner.hidden = !loading;
      }
    }

    function zbierzKontekstAi() {
      const seen = new Set();
      const out = [];

      function dodaj(nazwa) {
        const s = String(nazwa || '').trim();
        if (!s) return;
        const key = s.toLowerCase();
        if (seen.has(key)) return;
        seen.add(key);
        out.push(s);
      }

      wczytajKatalog().forEach(item => dodaj(item.nazwa));
      wczytajSzablonyPozycji().forEach(item => dodaj(item.nazwa));
      wczytajHistorie().slice(0, 20).forEach(wpis => {
        const pozycje = wpis && wpis.payload && Array.isArray(wpis.payload.pozycje) ? wpis.payload.pozycje : [];
        pozycje.forEach(p => dodaj(p && p.nazwa));
      });

      return out.slice(0, AI_MAX_KONTEKST);
    }

    function dopasujCenePozycji(nazwa, cena) {
      const parsed = Number(cena);
      if (Number.isFinite(parsed) && parsed > 0) {
        return { cena: parsed, zrodlo: null };
      }

      const key = String(nazwa || '').trim().toLowerCase();
      if (!key) return { cena: 0, zrodlo: null };

      const katalog = wczytajKatalog();
      const exactKat = katalog.find(k => k.nazwa.toLowerCase() === key);
      if (exactKat) {
        return { cena: exactKat.cena_jednostkowa, zrodlo: 'katalog' };
      }

      const szablony = wczytajSzablonyPozycji();
      const exactSz = szablony.find(s => s.nazwa.toLowerCase() === key);
      if (exactSz) {
        return { cena: exactSz.cena, zrodlo: 'szablon' };
      }

      const partial = katalog.filter(k => {
        const katKey = k.nazwa.toLowerCase();
        return katKey.includes(key) || key.includes(katKey);
      });
      if (partial.length === 1) {
        return { cena: partial[0].cena_jednostkowa, zrodlo: 'katalog' };
      }

      return { cena: 0, zrodlo: null };
    }

    function formatujCeneAi(cena) {
      const n = Number(cena);
      if (!Number.isFinite(n)) return '0,00';
      return n.toFixed(2).replace('.', ',');
    }

    function przygotujPozycjeAi(items) {
      if (!Array.isArray(items)) return [];
      return items
        .filter(item => item && typeof item === 'object')
        .map(item => {
          const nazwa = typeof item.nazwa === 'string' ? item.nazwa.trim() : '';
          const ilosc = Number(item.ilosc);
          const dopasowanie = dopasujCenePozycji(nazwa, item.cena);
          return {
            nazwa,
            ilosc,
            cena: dopasowanie.cena,
            zrodloCeny: dopasowanie.zrodlo,
          };
        })
        .filter(item => item.nazwa && Number.isFinite(item.ilosc) && item.ilosc > 0 && item.cena >= 0);
    }

    function wyczyscZdjecieAi() {
      if (aiPhotoPayload && aiPhotoPayload.previewUrl && aiPhotoPayload.previewUrl.startsWith('blob:')) {
        try { URL.revokeObjectURL(aiPhotoPayload.previewUrl); } catch (e) {}
      }
      aiPhotoPayload = null;
      if (aiPhotoInput) aiPhotoInput.value = '';
      if (aiPhotoThumb) aiPhotoThumb.classList.add('hidden');
      if (aiPhotoThumbImg) {
        aiPhotoThumbImg.removeAttribute('src');
        aiPhotoThumbImg.alt = '';
      }
    }

    function pokazMiniatureZdjeciaAi(previewUrl) {
      if (!aiPhotoThumb || !aiPhotoThumbImg) return;
      aiPhotoThumbImg.src = previewUrl;
      aiPhotoThumbImg.alt = 'Wybrane zdjęcie notatki';
      aiPhotoThumb.classList.remove('hidden');
    }

    function zamknijPodgladAi() {
      if (!modalAiPreview) return;
      modalAiPreview.classList.add('hidden');
      modalAiPreview.setAttribute('aria-hidden', 'true');
      document.body.style.overflow = '';
      aiPendingItems = [];
      if (aiPreviewTbody) aiPreviewTbody.innerHTML = '';
      if (aiPreviewMeta) aiPreviewMeta.textContent = '';
    }

    function pokazPodgladAi(items) {
      if (typeof window.zamknijAiInputSheet === 'function') {
        window.zamknijAiInputSheet();
      }
      if (!modalAiPreview || !aiPreviewTbody) return;
      aiPendingItems = items.slice();

      aiPreviewTbody.innerHTML = '';
      let uzupelnione = 0;
      aiPendingItems.forEach(item => {
        const tr = document.createElement('tr');

        const tdNazwa = document.createElement('td');
        tdNazwa.textContent = item.nazwa;
        if (item.zrodloCeny) {
          const tag = document.createElement('span');
          tag.className = 'ai-preview-tag';
          tag.textContent = item.zrodloCeny === 'katalog' ? 'z katalogu' : 'z historii';
          tdNazwa.appendChild(document.createTextNode(' '));
          tdNazwa.appendChild(tag);
          uzupelnione++;
        }

        const tdIlosc = document.createElement('td');
        tdIlosc.className = 'col-num tabular-nums';
        tdIlosc.textContent = String(item.ilosc);

        const tdCena = document.createElement('td');
        tdCena.className = 'col-num tabular-nums';
        tdCena.textContent = formatujCeneAi(item.cena);

        tr.appendChild(tdNazwa);
        tr.appendChild(tdIlosc);
        tr.appendChild(tdCena);
        aiPreviewTbody.appendChild(tr);
      });

      if (aiPreviewMeta) {
        let meta = 'Znaleziono ' + aiPendingItems.length + ' pozycji — sprawdź przed dodaniem.';
        if (uzupelnione > 0) {
          meta += ' Ceny uzupełnione z katalogu/historii: ' + uzupelnione + '.';
        }
        aiPreviewMeta.textContent = meta;
      }

      modalAiPreview.classList.remove('hidden');
      modalAiPreview.setAttribute('aria-hidden', 'false');
      document.body.style.overflow = 'hidden';
      if (btnAiPreviewDodaj) btnAiPreviewDodaj.focus();
    }

    function wstrzyknijPozycjeZAi(items) {
      if (!Array.isArray(items) || items.length === 0) {
        ustawAiParseStatus('Nie znaleziono pozycji w notatce.', 'error');
        return 0;
      }

      let dodane = 0;
      for (let idx = items.length - 1; idx >= 0; idx--) {
        const item = items[idx];
        if (!item || typeof item !== 'object') continue;
        const nazwa = typeof item.nazwa === 'string' ? item.nazwa.trim() : '';
        const ilosc = Number(item.ilosc);
        const dopasowanie = dopasujCenePozycji(nazwa, item.cena);
        const cena = dopasowanie.cena;
        if (!nazwa || !Number.isFinite(ilosc) || ilosc <= 0 || !Number.isFinite(cena) || cena < 0) continue;

        dodajWiersz();
        const tr = tbody.firstElementChild;
        if (!tr) continue;

        const inNazwa = tr.querySelector('.in-nazwa');
        const inIlosc = tr.querySelector('.in-ilosc');
        const inCena = tr.querySelector('.in-cena');

        if (inNazwa) {
          inNazwa.value = nazwa;
          inNazwa.dispatchEvent(new Event('input', { bubbles: true }));
        }
        if (inIlosc) {
          inIlosc.value = String(ilosc);
          inIlosc.dispatchEvent(new Event('input', { bubbles: true }));
        }
        if (inCena) {
          inCena.value = String(cena);
          inCena.dispatchEvent(new Event('input', { bubbles: true }));
        }
        dodane++;
      }

      if (dodane === 0) {
        ustawAiParseStatus('Nie znaleziono poprawnych pozycji do dodania.', 'error');
        return 0;
      }

      aktualizujSzacowanyZysk();
      saveDraft();
      return dodane;
    }

    function przygotujObrazDoAi(file) {
      return new Promise((resolve, reject) => {
        if (!file) {
          reject(new Error('empty'));
          return;
        }
        if (file.size > AI_PHOTO_MAX_BYTES) {
          reject(new Error('too_large'));
          return;
        }
        const mime = (file.type || '').toLowerCase();
        if (mime !== 'image/jpeg' && mime !== 'image/png') {
          reject(new Error('bad_type'));
          return;
        }

        const objectUrl = URL.createObjectURL(file);
        const img = new Image();
        img.onload = () => {
          try {
            const maxDim = 1600;
            let w = img.naturalWidth || img.width;
            let h = img.naturalHeight || img.height;
            if (w > maxDim || h > maxDim) {
              if (w >= h) {
                h = Math.round(h * maxDim / w);
                w = maxDim;
              } else {
                w = Math.round(w * maxDim / h);
                h = maxDim;
              }
            }
            const canvas = document.createElement('canvas');
            canvas.width = w;
            canvas.height = h;
            const ctx = canvas.getContext('2d');
            if (!ctx) {
              reject(new Error('canvas'));
              return;
            }
            ctx.drawImage(img, 0, 0, w, h);
            canvas.toBlob(blob => {
              URL.revokeObjectURL(objectUrl);
              if (!blob) {
                reject(new Error('blob'));
                return;
              }
              const reader = new FileReader();
              reader.onload = () => {
                const dataUrl = String(reader.result || '');
                const comma = dataUrl.indexOf(',');
                if (comma < 0) {
                  reject(new Error('dataurl'));
                  return;
                }
                resolve({
                  base64: dataUrl.slice(comma + 1),
                  mime_type: 'image/jpeg',
                  previewUrl: dataUrl,
                });
              };
              reader.onerror = () => reject(new Error('read'));
              reader.readAsDataURL(blob);
            }, 'image/jpeg', 0.82);
          } catch (e) {
            URL.revokeObjectURL(objectUrl);
            reject(e);
          }
        };
        img.onerror = () => {
          URL.revokeObjectURL(objectUrl);
          reject(new Error('image'));
        };
        img.src = objectUrl;
      });
    }

    async function przetworzNotatkeAi() {
      if (!btnAiParse) return;

      const tekst = aiNotatka ? aiNotatka.value.trim() : '';
      if (!tekst && !aiPhotoPayload) {
        ustawAiParseStatus('Dodaj tekst, dyktuj lub zdjęcie notatki przed przetwarzaniem.', 'error');
        return;
      }

      if (aiParseController) {
        aiParseController.abort();
      }
      aiParseController = new AbortController();

      ustawAiParseStatus('');
      ustawStanLadowaniaAiParse(true);

      const payload = { tekst, kontekst: zbierzKontekstAi() };
      if (aiPhotoPayload) {
        payload.obraz = aiPhotoPayload.base64;
        payload.mime_type = aiPhotoPayload.mime_type;
      }

      try {
        let retried429 = false;
        let resp = await fetch('/api/parse', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
          signal: aiParseController.signal,
        });

        if (resp.status === 429 && !retried429) {
          ustawAiParseStatus('Przeciążenie serwera, ponawiam...');
          await new Promise(function (resolve) { setTimeout(resolve, 2000); });
          retried429 = true;
          resp = await fetch('/api/parse', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
            signal: aiParseController.signal,
          });
        }

        let dane = null;
        try {
          dane = await resp.json();
        } catch (e) {
          dane = null;
        }

        if (!resp.ok) {
          const komunikat = (dane && dane.error)
            ? dane.error
            : (resp.status === 429
              ? 'Zbyt wiele żądań — spróbuj ponownie za chwilę.'
              : 'Nie udało się przetworzyć notatki.');
          ustawAiParseStatus(komunikat, 'error');
          return;
        }

        if (!Array.isArray(dane)) {
          ustawAiParseStatus('Nieprawidłowa odpowiedź serwera.', 'error');
          return;
        }

        const pozycje = przygotujPozycjeAi(dane);
        if (pozycje.length === 0) {
          ustawAiParseStatus('Nie znaleziono poprawnych pozycji w notatce.', 'error');
          return;
        }

        ustawAiParseStatus('Znaleziono ' + pozycje.length + ' pozycji — sprawdź przed dodaniem.', 'success');
        pokazPodgladAi(pozycje);
      } catch (err) {
        if (err && err.name === 'AbortError') return;
        ustawAiParseStatus('Błąd sieci — sprawdź połączenie i spróbuj ponownie.', 'error');
      } finally {
        ustawStanLadowaniaAiParse(false);
        aiParseController = null;
      }
    }

    function zatrzymajDyktowanie() {
      if (!speechRecognition || !speechRecording) return;
      speechRecording = false;
      try {
        speechRecognition.stop();
      } catch (e) {}
      if (btnAiMic) btnAiMic.classList.remove('is-recording');
    }

    function initAiSpeech() {
      const SpeechRecognitionCtor = window.SpeechRecognition || window.webkitSpeechRecognition;
      if (!SpeechRecognitionCtor || !btnAiMic) {
        if (btnAiMic) btnAiMic.style.display = 'none';
        return;
      }

      speechRecognition = new SpeechRecognitionCtor();
      speechRecognition.lang = 'pl-PL';
      speechRecognition.continuous = true;
      speechRecognition.interimResults = true;

      speechRecognition.addEventListener('result', (e) => {
        if (!aiNotatka) return;
        let finalTranscript = '';
        let interimTranscript = '';
        for (let i = e.resultIndex; i < e.results.length; i++) {
          const chunk = e.results[i][0].transcript;
          if (e.results[i].isFinal) finalTranscript += chunk;
          else interimTranscript += chunk;
        }
        if (finalTranscript) {
          speechBaseText += finalTranscript;
        }
        aiNotatka.value = speechBaseText + interimTranscript;
      });

      speechRecognition.addEventListener('end', () => {
        if (speechRecording) {
          try {
            speechRecognition.start();
          } catch (e) {
            zatrzymajDyktowanie();
          }
        } else if (btnAiMic) {
          btnAiMic.classList.remove('is-recording');
        }
      });

      speechRecognition.addEventListener('error', () => {
        zatrzymajDyktowanie();
        ustawAiParseStatus('Dyktowanie niedostępne w tej przeglądarce.', 'error');
      });

      btnAiMic.addEventListener('click', () => {
        if (speechRecording) {
          zatrzymajDyktowanie();
          return;
        }
        try {
          speechRecording = true;
          speechBaseText = aiNotatka ? aiNotatka.value : '';
          btnAiMic.classList.add('is-recording');
          speechRecognition.start();
        } catch (e) {
          zatrzymajDyktowanie();
          ustawAiParseStatus('Nie udało się uruchomić mikrofonu.', 'error');
        }
      });
    }

    function initAiPhoto() {
      if (btnAiPhoto && aiPhotoInput) {
        btnAiPhoto.addEventListener('click', () => aiPhotoInput.click());
        aiPhotoInput.addEventListener('change', async () => {
          const file = aiPhotoInput.files && aiPhotoInput.files[0];
          if (!file) return;
          ustawAiParseStatus('');
          try {
            wyczyscZdjecieAi();
            const prepared = await przygotujObrazDoAi(file);
            aiPhotoPayload = prepared;
            pokazMiniatureZdjeciaAi(prepared.previewUrl);
          } catch (e) {
            wyczyscZdjecieAi();
            if (e && e.message === 'too_large') {
              ustawAiParseStatus('Zdjęcie jest zbyt duże (max 5 MB przed kompresją).', 'error');
            } else if (e && e.message === 'bad_type') {
              ustawAiParseStatus('Wybierz plik JPG lub PNG.', 'error');
            } else {
              ustawAiParseStatus('Nie udało się wczytać zdjęcia.', 'error');
            }
          }
        });
      }
      if (btnAiPhotoRemove) {
        btnAiPhotoRemove.addEventListener('click', () => {
          wyczyscZdjecieAi();
          ustawAiParseStatus('');
        });
      }
    }

    function initAiPreviewModal() {
      if (btnAiPreviewAnuluj) {
        btnAiPreviewAnuluj.addEventListener('click', zamknijPodgladAi);
      }
      if (modalAiPreviewBackdrop) {
        modalAiPreviewBackdrop.addEventListener('click', zamknijPodgladAi);
      }
      if (btnAiPreviewDodaj) {
        btnAiPreviewDodaj.addEventListener('click', () => {
          const dodane = wstrzyknijPozycjeZAi(aiPendingItems);
          if (dodane > 0) {
            ustawAiParseStatus('Dodano ' + dodane + ' pozycji do wyceny.', 'success');
            if (aiNotatka) aiNotatka.value = '';
            wyczyscZdjecieAi();
          }
          zamknijPodgladAi();
        });
      }
      document.addEventListener('keydown', (e) => {
        if (e.key !== 'Escape') return;
        if (!modalAiPreview || modalAiPreview.classList.contains('hidden')) return;
        zamknijPodgladAi();
      });
    }

    function initAiQuickInput() {
      if (btnAiParse) {
        btnAiParse.addEventListener('click', przetworzNotatkeAi);
      }
      initAiPhoto();
      initAiPreviewModal();
      initAiSpeech();
    }

    function parsujCSV(text) {
      const rows = [];
      const src = text.replace(/^\uFEFF/, '');
      let cur = '';
      let row = [];
      let inQuotes = false;
      for (let i = 0; i < src.length; i++) {
        const ch = src[i];
        if (inQuotes) {
          if (ch === '"') {
            if (src[i + 1] === '"') { cur += '"'; i++; }
            else { inQuotes = false; }
          } else {
            cur += ch;
          }
        } else {
          if (ch === '"') {
            inQuotes = true;
          } else if (ch === ',' || ch === ';') {
            row.push(cur); cur = '';
          } else if (ch === '\n' || ch === '\r') {
            if (ch === '\r' && src[i + 1] === '\n') i++;
            row.push(cur); cur = '';
            if (row.some(f => f.trim() !== '')) rows.push(row.map(f => f.trim()));
            row = [];
          } else {
            cur += ch;
          }
        }
      }
      if (cur !== '' || row.length > 0) {
        row.push(cur);
        if (row.some(f => f.trim() !== '')) rows.push(row.map(f => f.trim()));
      }
      return rows;
    }

    function importujCSV(text) {
      const rows = parsujCSV(text);
      if (rows.length === 0) throw new Error('Plik jest pusty.');
      let startIdx = 0;
      const pierwsza = rows[0].map(f => f.toLowerCase());
      if (pierwsza[0] === 'nazwa' || pierwsza[0] === 'name') startIdx = 1;

      const items = [];
      for (let i = startIdx; i < rows.length; i++) {
        const r = rows[i];
        const nazwa = r[0];
        if (!nazwa) continue;
        const cenaRaw = (r[1] || '0').replace(/\s/g, '').replace(',', '.');
        const cena = parseFloat(cenaRaw);
        const kategoria = r[2] || 'Inne';
        items.push({
          nazwa,
          cena_jednostkowa: Number.isFinite(cena) ? cena : 0,
          kategoria,
        });
      }
      if (items.length === 0) throw new Error('Brak prawidłowych pozycji w pliku.');
      zapiszKatalog(items);
      return items;
    }

    const KOMUNIKAT_BLEDU_IMPORTU = 'Nie udało się odczytać pliku. Upewnij się, że zapisałeś go w Excelu jako CSV (rozdzielany przecinkami).';

    function escCSV(wartosc) {
      const s = String(wartosc == null ? '' : wartosc);
      if (/[",\r\n]/.test(s)) {
        return '"' + s.replace(/"/g, '""') + '"';
      }
      return s;
    }

    function eksportujKatalogCSV() {
      const items = wczytajKatalog();
      if (!items || items.length === 0) {
        pokazFeedbackKatalog('Katalog jest pusty — nie ma czego eksportować.', 'error');
        return;
      }

      const linie = ['nazwa,cena_jednostkowa,kategoria'];
      items.forEach(it => {
        const cena = Number(it.cena_jednostkowa);
        const cenaTxt = Number.isFinite(cena) ? cena.toFixed(2) : '0.00';
        linie.push(
          escCSV(it.nazwa) + ',' +
          escCSV(cenaTxt) + ',' +
          escCSV(it.kategoria || 'Inne')
        );
      });

      const csv = '\uFEFF' + linie.join('\r\n') + '\r\n';
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'moje_uslugi_sumit.csv';
      document.body.appendChild(a);
      a.click();
      a.remove();
      setTimeout(() => URL.revokeObjectURL(url), 0);

      pokazFeedbackKatalog('Wyeksportowano ' + items.length + ' pozycji do pliku CSV.', 'success');
    }

    btnExportCatalog.addEventListener('click', eksportujKatalogCSV);

    btnImportCatalog.addEventListener('click', () => importCatalogInput.click());

    importCatalogInput.addEventListener('change', () => {
      const file = importCatalogInput.files && importCatalogInput.files[0];
      if (!file) return;
      const reader = new FileReader();
      reader.onload = () => {
        try {
          const items = importujCSV(String(reader.result || ''));
          aktywnaKategoria = '';
          catalogSearch.value = '';
          renderujKatalog();
          pokazFeedbackKatalog('Zaimportowano ' + items.length + ' pozycji z pliku Excela.', 'success');
        } catch (err) {
          console.warn('Import katalogu:', err);
          pokazFeedbackKatalog(KOMUNIKAT_BLEDU_IMPORTU, 'error');
        } finally {
          importCatalogInput.value = '';
        }
      };
      reader.onerror = () => {
        pokazFeedbackKatalog(KOMUNIKAT_BLEDU_IMPORTU, 'error');
        importCatalogInput.value = '';
      };
      reader.readAsText(file, 'utf-8');
    });

    function otworzKatalog() {
      if (!aktywnyWiersz || !tbody.contains(aktywnyWiersz)) {
        aktywnyWiersz = tbody.lastElementChild;
      }
      catalogSearch.value = '';
      aktywnaKategoria = '';
      ukryjFeedbackKatalog();
      renderujKatalog();
      catalogModal.hidden = false;
      catalogModal.setAttribute('aria-hidden', 'false');
      document.body.style.overflow = 'hidden';
      setTimeout(() => catalogSearch.focus(), 0);
    }

    function zamknijKatalog() {
      catalogModal.hidden = true;
      catalogModal.setAttribute('aria-hidden', 'true');
      document.body.style.overflow = '';
    }

    btnKatalog.addEventListener('click', otworzKatalog);
    btnZamknijCatalog.addEventListener('click', zamknijKatalog);
    catalogBackdrop.addEventListener('click', zamknijKatalog);
    catalogSearch.addEventListener('input', () => renderujKatalog());
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && !catalogModal.hidden) zamknijKatalog();
    });

    const STORAGE_KEY_HISTORIA = 'sumit_historia';
    const MAX_HISTORIA = 100;

    const historiaModal = document.getElementById('historia-modal');
    const historiaBackdrop = document.getElementById('historia-modal-backdrop');
    const historiaList = document.getElementById('historia-list');
    const btnHistoria = document.getElementById('btn-historia');
    const btnZamknijHistoria = document.getElementById('btn-zamknij-historia');
    const btnHistoriaWyczysc = document.getElementById('btn-historia-wyczysc');

    function wczytajHistorie() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY_HISTORIA);
        if (!raw) return [];
        const dane = JSON.parse(raw);
        if (!Array.isArray(dane)) return [];
        return dane.filter(w => w && typeof w === 'object' && w.payload);
      } catch (e) {
        return [];
      }
    }

    function zapiszHistorie(items) {
      try {
        localStorage.setItem(STORAGE_KEY_HISTORIA, JSON.stringify(items));
      } catch (e) {}
    }

    function obliczSumePozycji(pozycje) {
      if (!Array.isArray(pozycje)) return 0;
      let suma = 0;
      pozycje.forEach(p => {
        const ilosc = Number(p && p.ilosc);
        const cena = Number(p && p.cena_jednostkowa);
        if (Number.isFinite(ilosc) && Number.isFinite(cena)) {
          suma += ilosc * cena;
        }
      });
      return Math.round(suma * 100) / 100;
    }

    // Zwraca { zysk, kompletny } gdzie:
    //   zysk      — suma (cena - koszt) * ilosc dla pozycji z kompletem danych,
    //   kompletny — true, gdy KAŻDA pozycja ma poprawny koszt; w przeciwnym razie
    //               wpis traktujemy jako "częściowy" i pomijamy go w sumach statystyk.
    function obliczZyskWpisu(pozycje, koszty) {
      if (!Array.isArray(pozycje) || pozycje.length === 0) return { zysk: 0, kompletny: false };
      if (!Array.isArray(koszty)) return { zysk: 0, kompletny: false };
      let zysk = 0;
      let kompletny = true;
      pozycje.forEach((p, i) => {
        const ilosc = Number(p && p.ilosc);
        const cena  = Number(p && p.cena_jednostkowa);
        const koszt = Number(koszty[i]);
        if (Number.isFinite(ilosc) && Number.isFinite(cena) && Number.isFinite(koszt)) {
          zysk += (cena - koszt) * ilosc;
        } else {
          kompletny = false;
        }
      });
      return { zysk: Math.round(zysk * 100) / 100, kompletny };
    }

    function dodajDoHistorii(payload, koszty, token) {
      if (!payload || typeof payload !== 'object') return;
      const koszt = Array.isArray(koszty) ? koszty.slice() : null;
      const { zysk, kompletny } = obliczZyskWpisu(payload.pozycje, koszt || []);
      const wpis = {
        id: Date.now(),
        dataZapisu: new Date().toISOString(),
        numerOferty: String(payload.numer_oferty || ''),
        klient: String(payload.klient || ''),
        suma: obliczSumePozycji(payload.pozycje),
        // koszty i zysk są zapisywane WYŁĄCZNIE w localStorage — nie w payloadzie /quote.
        koszty: koszt,
        zysk: kompletny ? zysk : null,
        token: token || null,
        akceptacja: null,
        payload: payload,
      };
      const lista = wczytajHistorie();
      lista.unshift(wpis);
      while (lista.length > MAX_HISTORIA) lista.pop();
      zapiszHistorie(lista);
      zapiszSzablonyZPayloadu(payload);
      upsertKlientaZTekstu(payload.klient || '');
    }

    const STORAGE_KEY_KLIENCI = 'sumit_klienci';
    const MAX_KLIENCI = 100;

    function _canonNip(v) {
      return String(v || '').replace(/\D/g, '').slice(0, 10);
    }

    function _canonNazwa(v) {
      return String(v || '').trim().toLowerCase();
    }

    function wczytajKlientow() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY_KLIENCI);
        if (!raw) return [];
        const dane = JSON.parse(raw);
        if (!Array.isArray(dane)) return [];
        return dane
          .filter(k => k && typeof k === 'object')
          .map(k => ({
            nip: String(k.nip || '').trim(),
            nazwa: String(k.nazwa || '').trim(),
            adres: String(k.adres || '').trim(),
          }))
          .filter(k => k.nip || k.nazwa);
      } catch (e) {
        return [];
      }
    }

    function zapiszKlientow(items) {
      try {
        localStorage.setItem(STORAGE_KEY_KLIENCI, JSON.stringify(items));
      } catch (e) {}
    }

    function parseKlientBlok(text) {
      const linie = String(text || '').split('\n').map(s => s.trim()).filter(Boolean);
      let nip = '';
      const reszta = [];
      linie.forEach(l => {
        if (!nip) {
          const m = l.match(/^NIP[\s:.-]*([\d\s-]+)$/i);
          if (m) {
            const cyfry = m[1].replace(/\D/g, '');
            if (cyfry.length === 10) { nip = cyfry; return; }
          }
        }
        reszta.push(l);
      });
      const nazwa = reszta.shift() || '';
      const adres = reszta.join('\n');
      return { nip, nazwa, adres };
    }

    function formatKlientBlok(rekord) {
      if (!rekord) return '';
      const linie = [];
      if (rekord.nazwa) linie.push(String(rekord.nazwa));
      if (rekord.adres) linie.push(String(rekord.adres));
      if (rekord.nip)   linie.push('NIP: ' + String(rekord.nip));
      return linie.join('\n');
    }

    function upsertKlienta(rekord) {
      if (!rekord) return;
      const r = {
        nip: _canonNip(rekord.nip),
        nazwa: String(rekord.nazwa || '').trim(),
        adres: String(rekord.adres || '').trim(),
      };
      if (!r.nazwa && !r.nip) return;

      const lista = wczytajKlientow();
      let idx = -1;
      if (r.nip) {
        idx = lista.findIndex(k => _canonNip(k.nip) === r.nip);
      }
      if (idx === -1 && r.nazwa) {
        const nz = _canonNazwa(r.nazwa);
        idx = lista.findIndex(k => _canonNazwa(k.nazwa) === nz);
      }

      let scalony = r;
      if (idx !== -1) {
        const stary = lista[idx];
        scalony = {
          nip:   r.nip   || stary.nip   || '',
          nazwa: r.nazwa || stary.nazwa || '',
          adres: r.adres || stary.adres || '',
        };
        lista.splice(idx, 1);
      }
      lista.unshift(scalony);
      while (lista.length > MAX_KLIENCI) lista.pop();
      zapiszKlientow(lista);
    }

    function upsertKlientaZTekstu(text) {
      const parsed = parseKlientBlok(text);
      if (!parsed.nazwa && !parsed.nip) return;
      upsertKlienta(parsed);
    }

    function znajdzKlientaPoZapytaniu(zapytanie) {
      const raw = String(zapytanie || '').trim();
      if (!raw) return null;
      const lista = wczytajKlientow();
      if (lista.length === 0) return null;
      const cyfry = raw.replace(/\D/g, '');
      if (cyfry.length === 10) {
        const trafioneNip = lista.find(k => _canonNip(k.nip) === cyfry);
        if (trafioneNip) return trafioneNip;
      }
      const nz = raw.toLowerCase();
      const trafioneNazwa = lista.find(k => k.nazwa && k.nazwa.toLowerCase() === nz);
      return trafioneNazwa || null;
    }

    const STORAGE_KEY_SZABLONY = 'sumit_szablony_pozycji';
    const MAX_SZABLONY_POZYCJI = 200;
    let _szablonyPozycjiCache = null;

    function wczytajSzablonyPozycji() {
      if (_szablonyPozycjiCache) return _szablonyPozycjiCache;
      try {
        const raw = localStorage.getItem(STORAGE_KEY_SZABLONY);
        if (!raw) { _szablonyPozycjiCache = []; return _szablonyPozycjiCache; }
        const dane = JSON.parse(raw);
        if (!Array.isArray(dane)) { _szablonyPozycjiCache = []; return _szablonyPozycjiCache; }
        _szablonyPozycjiCache = dane
          .filter(s => s && typeof s === 'object' && typeof s.nazwa === 'string' && s.nazwa.trim())
          .map(s => ({
            nazwa: String(s.nazwa).trim(),
            jednostka: typeof s.jednostka === 'string' ? s.jednostka : '',
            cena: Number(s.cena),
          }))
          .filter(s => Number.isFinite(s.cena) && s.cena >= 0);
        return _szablonyPozycjiCache;
      } catch (e) {
        _szablonyPozycjiCache = [];
        return _szablonyPozycjiCache;
      }
    }

    function zapiszSzablonyPozycji(items) {
      try {
        localStorage.setItem(STORAGE_KEY_SZABLONY, JSON.stringify(items));
        _szablonyPozycjiCache = items.slice();
      } catch (e) {}
    }

    function zapiszSzablonyZPayloadu(payload) {
      if (!payload || !Array.isArray(payload.pozycje)) return;
      const mapa = new Map();
      wczytajSzablonyPozycji().forEach(it => mapa.set(it.nazwa.toLowerCase(), it));

      let zmienione = false;
      payload.pozycje.forEach(p => {
        const nazwa = String((p && p.nazwa) || '').trim();
        const cena = Number(p && p.cena_jednostkowa);
        if (!nazwa || !Number.isFinite(cena) || cena < 0) return;
        const klucz = nazwa.toLowerCase();
        const wpis = { nazwa, jednostka: '', cena: Math.round(cena * 100) / 100 };
        mapa.delete(klucz);
        mapa.set(klucz, wpis);
        zmienione = true;
      });

      if (!zmienione) return;
      const noweKolejne = Array.from(mapa.values()).reverse().slice(0, MAX_SZABLONY_POZYCJI);
      zapiszSzablonyPozycji(noweKolejne);
      odswiezDatalistSzablony();
    }

    function odswiezDatalistSzablony() {
      const datalist = document.getElementById('szablony-pozycji-datalist');
      if (!datalist) return;
      const szablony = wczytajSzablonyPozycji();
      datalist.innerHTML = '';
      const frag = document.createDocumentFragment();
      szablony.forEach(s => {
        const opt = document.createElement('option');
        opt.value = s.nazwa;
        const cenaFmt = (Number(s.cena) || 0).toFixed(2).replace('.', ',') + ' zł';
        opt.label = 'Ostatnio: ' + cenaFmt;
        frag.appendChild(opt);
      });
      datalist.appendChild(frag);
    }

    function aplikujSzablonDoWiersza(tr) {
      if (!tr) return;
      const nazwaInput = tr.querySelector('.in-nazwa');
      const cenaInput = tr.querySelector('.in-cena');
      if (!nazwaInput || !cenaInput) return;
      const wartosc = String(nazwaInput.value || '').trim();
      if (!wartosc) {
        tr.dataset.szablonNazwa = '';
        return;
      }
      const klucz = wartosc.toLowerCase();
      const szablon = wczytajSzablonyPozycji().find(s => s.nazwa.toLowerCase() === klucz);
      if (!szablon) {
        tr.dataset.szablonNazwa = '';
        return;
      }
      if (tr.dataset.szablonNazwa === szablon.nazwa) return;
      cenaInput.value = String(szablon.cena);
      cenaInput.dispatchEvent(new Event('input', { bubbles: true }));
      tr.dataset.szablonNazwa = szablon.nazwa;
    }

    odswiezDatalistSzablony();

    function usunZHistorii(id) {
      const lista = wczytajHistorie().filter(w => w.id !== id);
      zapiszHistorie(lista);
    }

    function formatujDateZapisu(iso) {
      const d = new Date(iso);
      if (Number.isNaN(d.getTime())) return String(iso || '');
      const dd = String(d.getDate()).padStart(2, '0');
      const mm = String(d.getMonth() + 1).padStart(2, '0');
      const yyyy = d.getFullYear();
      const hh = String(d.getHours()).padStart(2, '0');
      const min = String(d.getMinutes()).padStart(2, '0');
      return `${dd}.${mm}.${yyyy}, ${hh}:${min}`;
    }

    function formatujSume(suma) {
      const n = Number(suma);
      if (!Number.isFinite(n)) return '0,00 zł';
      return n.toFixed(2).replace('.', ',') + ' zł';
    }

    const statsSection = document.getElementById('stats-section');

    function tworzKarteStat(label, value, meta, modCls) {
      const card = document.createElement('div');
      card.className = 'stats-card' + (modCls ? ' ' + modCls : '');
      const l = document.createElement('div');
      l.className = 'stats-card-label';
      l.textContent = label;
      card.appendChild(l);
      if (value != null) {
        const v = document.createElement('div');
        v.className = 'stats-card-value';
        v.textContent = value;
        card.appendChild(v);
      }
      if (meta != null) {
        const m = document.createElement('div');
        m.className = 'stats-card-meta';
        m.textContent = meta;
        card.appendChild(m);
      }
      return card;
    }

    const NAZWY_MIESIECY = ['Sty', 'Lut', 'Mar', 'Kwi', 'Maj', 'Cze', 'Lip', 'Sie', 'Wrz', 'Paź', 'Lis', 'Gru'];
    const NAZWY_DNI = ['Niedziela', 'Poniedziałek', 'Wtorek', 'Środa', 'Czwartek', 'Piątek', 'Sobota'];
    const KOLEJNOSC_DNI_TYGODNIA = [1, 2, 3, 4, 5, 6, 0];

    const STATS_OKRES_KEY = 'sumit_stats_okres';
    const STATS_OKRES_DEFAULT = '30';
    const STATS_OKRESY = [
      { id: '7', label: '7 dni', dni: 7 },
      { id: '30', label: '30 dni', dni: 30 },
      { id: '90', label: '90 dni', dni: 90 },
      { id: '365', label: 'Rok', dni: 365 },
      { id: 'all', label: 'Wszystko', dni: null }
    ];

    function wczytajOkresStat() {
      try {
        const v = localStorage.getItem(STATS_OKRES_KEY);
        if (STATS_OKRESY.some(o => o.id === v)) return v;
      } catch (e) {}
      return STATS_OKRES_DEFAULT;
    }

    function zapiszOkresStat(okres) {
      try { localStorage.setItem(STATS_OKRES_KEY, okres); } catch (e) {}
    }

    function initAppSettingsStatsChips() {
      const wrap = document.getElementById('app-settings-stats-okres');
      if (!wrap || wrap.dataset.bound === '1') return;
      wrap.dataset.bound = '1';
      wrap.addEventListener('click', (e) => {
        const chip = e.target.closest('.chip[data-stats-okres]');
        if (!chip) return;
        const okres = chip.getAttribute('data-stats-okres');
        if (!STATS_OKRESY.some((o) => o.id === okres)) return;
        zapiszOkresStat(okres);
        odswiezChipyStatsOkresApp();
        const widokStat = document.getElementById('view-statystyki');
        if (widokStat && !widokStat.classList.contains('hidden') && !widokStat.hasAttribute('hidden')) {
          renderStatystyki(okres);
        }
      });
    }

    function filtrujHistoriePoOkresie(lista, okres) {
      if (!Array.isArray(lista)) return [];
      const def = STATS_OKRESY.find(o => o.id === okres);
      if (!def || def.dni == null) return lista.slice();
      const odKiedy = Date.now() - def.dni * 24 * 60 * 60 * 1000;
      return lista.filter(wpis => {
        const data = new Date(wpis.dataZapisu);
        if (Number.isNaN(data.getTime())) return false;
        return data.getTime() >= odKiedy;
      });
    }

    function policzTopDzienTygodnia(lista) {
      const liczniki = [0, 0, 0, 0, 0, 0, 0];
      lista.forEach(wpis => {
        const data = new Date(wpis.dataZapisu);
        if (Number.isNaN(data.getTime())) return;
        liczniki[data.getDay()] += 1;
      });
      let najlepszy = -1;
      let najwiecej = 0;
      KOLEJNOSC_DNI_TYGODNIA.forEach(d => {
        if (liczniki[d] > najwiecej) {
          najwiecej = liczniki[d];
          najlepszy = d;
        }
      });
      return { indeks: najlepszy, liczba: najwiecej };
    }

    function sufiksOferty(n) {
      if (n === 1) return 'wycena';
      if (n >= 2 && n <= 4) return 'wyceny';
      return 'wycen';
    }

    let chartTooltip = null;
    let heatmapTooltip = null;

    function ukryjTooltipHeatmapy() {
      if (heatmapTooltip) heatmapTooltip.classList.remove('is-visible');
    }

    function pokazTooltipHeatmapy(cell) {
      const card = cell.closest('.heatmap-card');
      if (!card) return;
      if (!heatmapTooltip || heatmapTooltip.parentElement !== card) {
        if (heatmapTooltip && heatmapTooltip.parentElement) {
          heatmapTooltip.parentElement.removeChild(heatmapTooltip);
        }
        heatmapTooltip = document.createElement('div');
        heatmapTooltip.className = 'chart-tooltip heatmap-tooltip';
        heatmapTooltip.setAttribute('role', 'tooltip');
        card.appendChild(heatmapTooltip);
      }

      const label = cell.dataset.label || '';
      const liczba = parseInt(cell.dataset.count || '0', 10) || 0;

      heatmapTooltip.innerHTML = '';
      const t = document.createElement('div');
      t.className = 'chart-tooltip-title';
      t.textContent = label;
      const c = document.createElement('div');
      c.className = 'chart-tooltip-amount';
      c.textContent = liczba + ' ' + sufiksOferty(liczba);
      heatmapTooltip.appendChild(t);
      heatmapTooltip.appendChild(c);

      const cellRect = cell.getBoundingClientRect();
      const cardRect = card.getBoundingClientRect();
      heatmapTooltip.style.left = (cellRect.left - cardRect.left + cellRect.width / 2) + 'px';
      heatmapTooltip.style.top = (cellRect.top - cardRect.top - 8) + 'px';
      heatmapTooltip.style.transform = 'translate(-50%, -100%)';
      heatmapTooltip.classList.add('is-visible');
    }

    function ukryjTooltipsStatystyk() {
      ukryjTooltipWykresu();
      ukryjTooltipHeatmapy();
    }

    function pokazTooltipWykresu(bar) {
      const chartContainer = bar.closest('.chart-container');
      if (!chartContainer) return;
      if (!chartTooltip || chartTooltip.parentElement !== chartContainer) {
        if (chartTooltip && chartTooltip.parentElement) {
          chartTooltip.parentElement.removeChild(chartTooltip);
        }
        chartTooltip = document.createElement('div');
        chartTooltip.className = 'chart-tooltip';
        chartContainer.appendChild(chartTooltip);
      }
      const label = bar.dataset.label || '';
      const suma = Number(bar.dataset.suma) || 0;
      const liczba = parseInt(bar.dataset.liczba || '0', 10) || 0;

      chartTooltip.innerHTML = '';
      const t = document.createElement('div');
      t.className = 'chart-tooltip-title';
      t.textContent = label;
      const a = document.createElement('div');
      a.className = 'chart-tooltip-amount';
      a.textContent = formatujSume(suma);
      const c = document.createElement('div');
      c.className = 'chart-tooltip-count';
      c.textContent = liczba + ' ' + sufiksOferty(liczba);
      chartTooltip.appendChild(t);
      chartTooltip.appendChild(a);
      chartTooltip.appendChild(c);

      const barRect = bar.getBoundingClientRect();
      const containerRect = chartContainer.getBoundingClientRect();
      chartTooltip.style.left = (barRect.left - containerRect.left + barRect.width / 2) + 'px';
      chartTooltip.style.top = (barRect.top - containerRect.top - 8) + 'px';
      chartTooltip.style.transform = 'translate(-50%, -100%)';
      chartTooltip.classList.add('is-visible');
    }

    function ukryjTooltipWykresu() {
      if (chartTooltip) chartTooltip.classList.remove('is-visible');
    }

    function agregujMiesiacami(lista, ileMiesiecy) {
      const teraz = new Date();
      const buckets = [];
      for (let i = ileMiesiecy - 1; i >= 0; i--) {
        const d = new Date(teraz.getFullYear(), teraz.getMonth() - i, 1);
        buckets.push({
          key: d.getFullYear() + '-' + d.getMonth(),
          label: NAZWY_MIESIECY[d.getMonth()],
          suma: 0,
          liczba: 0
        });
      }
      const mapa = new Map(buckets.map(b => [b.key, b]));
      lista.forEach(wpis => {
        const data = new Date(wpis.dataZapisu);
        if (Number.isNaN(data.getTime())) return;
        const key = data.getFullYear() + '-' + data.getMonth();
        const b = mapa.get(key);
        if (!b) return;
        b.suma += Number(wpis.suma) || 0;
        b.liczba += 1;
      });
      return buckets;
    }

    function renderujWykres(buckets) {
      const chart = document.getElementById('stat-chart');
      const labels = document.getElementById('stat-chart-labels');
      if (!chart || !labels) return;
      ukryjTooltipsStatystyk();
      chart.innerHTML = '';
      labels.innerHTML = '';
      const max = buckets.reduce((m, b) => Math.max(m, b.suma), 0);
      const slupki = [];
      buckets.forEach(b => {
        const bar = document.createElement('div');
        bar.className = 'chart-bar' + (b.suma === 0 ? ' is-empty' : '');
        const pct = max > 0 ? (b.suma / max) * 100 : 0;
        const docelowa = b.suma === 0 ? '2px' : Math.max(pct, 2) + '%';
        bar.dataset.docelowa = docelowa;
        bar.dataset.label = b.label;
        bar.dataset.suma = String(b.suma);
        bar.dataset.liczba = String(b.liczba);
        bar.style.height = '0%';
        bar.style.minHeight = '0';

        bar.addEventListener('mouseenter', () => pokazTooltipWykresu(bar));
        bar.addEventListener('mouseleave', ukryjTooltipWykresu);

        chart.appendChild(bar);
        slupki.push(bar);

        const lbl = document.createElement('div');
        lbl.textContent = b.label;
        labels.appendChild(lbl);
      });

      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          slupki.forEach(bar => {
            bar.style.height = bar.dataset.docelowa;
            bar.style.minHeight = '';
          });
        });
      });
    }

    function nazwaKlientaPrimary(s) {
      if (!s || typeof s !== 'string') return '';
      return s.split('\n')[0].trim();
    }

    function agregujKlientow(lista) {
      const mapa = new Map();
      lista.forEach(wpis => {
        const klient = nazwaKlientaPrimary(wpis && wpis.klient);
        if (!klient) return;
        const suma = Number(wpis.suma) || 0;
        const dataMs = (() => {
          const d = new Date(wpis.dataZapisu);
          return Number.isNaN(d.getTime()) ? 0 : d.getTime();
        })();
        let agg = mapa.get(klient);
        if (!agg) {
          agg = { klient: klient, liczba: 0, suma: 0, ostatniaMs: 0 };
          mapa.set(klient, agg);
        }
        agg.liczba += 1;
        agg.suma += suma;
        if (dataMs > agg.ostatniaMs) agg.ostatniaMs = dataMs;
      });
      return Array.from(mapa.values()).map(a => ({
        klient: a.klient,
        liczba: a.liczba,
        suma: a.suma,
        srednia: a.liczba > 0 ? a.suma / a.liczba : 0,
        ostatniaMs: a.ostatniaMs
      })).sort((a, b) => b.suma - a.suma);
    }

    function formatujDateKrotka(ms) {
      if (!ms) return '—';
      const d = new Date(ms);
      if (Number.isNaN(d.getTime())) return '—';
      const dd = String(d.getDate()).padStart(2, '0');
      const mm = String(d.getMonth() + 1).padStart(2, '0');
      return dd + '.' + mm + '.' + d.getFullYear();
    }

    function renderujKarteTopKlienci(lista, okresLabel) {
      const card = tworzKarteStat('Top klienci', null, null, 'stats-card-full stats-card-standalone');
      const klienci = agregujKlientow(lista).slice(0, 5);

      if (klienci.length === 0) {
        const empty = document.createElement('div');
        empty.className = 'stats-klienci-empty';
        empty.textContent = 'Brak danych klientów w wybranym okresie';
        card.appendChild(empty);
        return card;
      }

      const wrap = document.createElement('div');
      wrap.className = 'stats-klienci-wrap';

      const table = document.createElement('table');
      table.className = 'stats-klienci-table';

      const thead = document.createElement('thead');
      const trh = document.createElement('tr');
      [
        { txt: 'Klient', cls: '' },
        { txt: 'Liczba wycen', cls: 'col-num' },
        { txt: 'Łączna wartość', cls: 'col-num' },
        { txt: 'Średnia wartość', cls: 'col-num' },
        { txt: 'Ostatnia wycena', cls: 'col-num' }
      ].forEach(h => {
        const th = document.createElement('th');
        th.textContent = h.txt;
        if (h.cls) th.className = h.cls;
        trh.appendChild(th);
      });
      thead.appendChild(trh);
      table.appendChild(thead);

      const tbodyEl = document.createElement('tbody');
      klienci.forEach(k => {
        const tr = document.createElement('tr');

        const tdK = document.createElement('td');
        tdK.className = 'col-klient';
        tdK.textContent = k.klient;
        tdK.title = k.klient;
        tr.appendChild(tdK);

        const tdL = document.createElement('td');
        tdL.className = 'col-num';
        tdL.textContent = String(k.liczba);
        tr.appendChild(tdL);

        const tdS = document.createElement('td');
        tdS.className = 'col-num';
        tdS.textContent = formatujSume(k.suma);
        tr.appendChild(tdS);

        const tdA = document.createElement('td');
        tdA.className = 'col-num';
        tdA.textContent = formatujSume(k.srednia);
        tr.appendChild(tdA);

        const tdD = document.createElement('td');
        tdD.className = 'col-num';
        tdD.textContent = formatujDateKrotka(k.ostatniaMs);
        tr.appendChild(tdD);

        tbodyEl.appendChild(tr);
      });
      table.appendChild(tbodyEl);
      wrap.appendChild(table);
      card.appendChild(wrap);
      return card;
    }

    const HEATMAPA_DNI_SKROTY = ['Nd', 'Pn', 'Wt', 'Śr', 'Cz', 'Pt', 'Sb'];
    const HEATMAPA_NAZWY_DNI = ['Niedziela', 'Poniedziałek', 'Wtorek', 'Środa', 'Czwartek', 'Piątek', 'Sobota'];
    const HEATMAPA_MIESIACE_PELNE = [
      'Sty', 'Lut', 'Mar', 'Kwi', 'Maj', 'Cze',
      'Lip', 'Sie', 'Wrz', 'Paź', 'Lis', 'Gru'
    ];

    function startTygodniaPnUTC(d) {
      const out = new Date(d.getFullYear(), d.getMonth(), d.getDate());
      const day = out.getDay();
      const offset = day === 0 ? 6 : day - 1;
      out.setDate(out.getDate() - offset);
      return out;
    }

    function dataKluczLokalna(d) {
      const yy = d.getFullYear();
      const mm = String(d.getMonth() + 1).padStart(2, '0');
      const dd = String(d.getDate()).padStart(2, '0');
      return yy + '-' + mm + '-' + dd;
    }

    function agregujDniHeatmapy(historiaPelna) {
      const mapa = new Map();
      historiaPelna.forEach(wpis => {
        const d = new Date(wpis.dataZapisu);
        if (Number.isNaN(d.getTime())) return;
        const klucz = dataKluczLokalna(d);
        mapa.set(klucz, (mapa.get(klucz) || 0) + 1);
      });
      return mapa;
    }

    function poziomHeatmapy(n) {
      if (n <= 0) return 0;
      if (n === 1) return 1;
      if (n <= 3) return 2;
      return 3;
    }

    function renderujKarteHeatmapy(historiaPelna) {
      const card = tworzKarteStat('Aktywność (ostatnie 52 tygodnie)', null, null, 'stats-card-full stats-card-standalone heatmap-card');

      const mapa = agregujDniHeatmapy(historiaPelna);

      const dzis = new Date();
      const dzisLokalne = new Date(dzis.getFullYear(), dzis.getMonth(), dzis.getDate());
      const dzisMs = dzisLokalne.getTime();
      const lastMon = startTygodniaPnUTC(dzisLokalne);
      const firstMon = new Date(lastMon.getFullYear(), lastMon.getMonth(), lastMon.getDate() - 51 * 7);

      const scroll = document.createElement('div');
      scroll.className = 'heatmap-scroll';

      const wrap = document.createElement('div');
      wrap.className = 'heatmap-wrap';

      const monthLabels = document.createElement('div');
      monthLabels.className = 'heatmap-month-labels';

      const dayLabels = document.createElement('div');
      dayLabels.className = 'heatmap-day-labels';
      [0, 1, 2, 3, 4, 5, 6].forEach(rowIdx => {
        const span = document.createElement('span');
        if (rowIdx === 0) span.textContent = 'Pn';
        else if (rowIdx === 2) span.textContent = 'Śr';
        else if (rowIdx === 4) span.textContent = 'Pt';
        else span.textContent = '';
        dayLabels.appendChild(span);
      });

      const grid = document.createElement('div');
      grid.className = 'heatmap-grid';

      let prevMonth = -1;
      for (let w = 0; w < 52; w++) {
        const tygStart = new Date(firstMon.getFullYear(), firstMon.getMonth(), firstMon.getDate() + w * 7);
        if (tygStart.getMonth() !== prevMonth) {
          const lbl = document.createElement('span');
          lbl.textContent = HEATMAPA_MIESIACE_PELNE[tygStart.getMonth()];
          lbl.style.setProperty('--hm-col', String(w));
          monthLabels.appendChild(lbl);
          prevMonth = tygStart.getMonth();
        }

        for (let d = 0; d < 7; d++) {
          const dzien = new Date(tygStart.getFullYear(), tygStart.getMonth(), tygStart.getDate() + d);
          const klucz = dataKluczLokalna(dzien);
          const liczba = mapa.get(klucz) || 0;
          const cell = document.createElement('div');
          const poz = poziomHeatmapy(liczba);
          cell.className = 'heatmap-cell' + (poz > 0 ? ' l' + poz : '');
          if (dzien.getTime() > dzisMs) cell.classList.add('is-future');

          const dataFmt = formatujDateKrotka(dzien.getTime());
          const dzienNazwa = HEATMAPA_NAZWY_DNI[dzien.getDay()];
          cell.classList.add('is-interactive');
          cell.dataset.count = String(liczba);
          cell.dataset.label = dzienNazwa + ', ' + dataFmt;
          cell.setAttribute('aria-label', dataFmt + ': ' + liczba + ' ' + sufiksOferty(liczba));
          cell.tabIndex = 0;
          cell.addEventListener('mouseenter', () => pokazTooltipHeatmapy(cell));
          cell.addEventListener('mouseleave', ukryjTooltipHeatmapy);
          cell.addEventListener('focus', () => pokazTooltipHeatmapy(cell));
          cell.addEventListener('blur', ukryjTooltipHeatmapy);
          grid.appendChild(cell);
        }
      }

      wrap.appendChild(monthLabels);
      wrap.appendChild(dayLabels);
      wrap.appendChild(grid);
      scroll.appendChild(wrap);
      card.appendChild(scroll);

      const legend = document.createElement('div');
      legend.className = 'heatmap-legend';
      const lblMniej = document.createElement('span');
      lblMniej.textContent = 'mniej';
      legend.appendChild(lblMniej);
      [0, 1, 2, 3].forEach(p => {
        const c = document.createElement('div');
        c.className = 'heatmap-cell' + (p > 0 ? ' l' + p : '');
        legend.appendChild(c);
      });
      const lblWiecej = document.createElement('span');
      lblWiecej.textContent = 'więcej';
      legend.appendChild(lblWiecej);
      card.appendChild(legend);

      return card;
    }

    function renderStatystyki(okresArg) {
      if (!statsSection) return;
      ukryjTooltipsStatystyk();
      const calaHistoria = wczytajHistorie();

      statsSection.innerHTML = '';
      const title = document.createElement('h2');
      title.className = 'section-title';
      title.textContent = 'Statystyki';
      statsSection.appendChild(title);

      if (!calaHistoria || calaHistoria.length === 0) {
        const empty = document.createElement('div');
        empty.className = 'stats-empty';
        empty.textContent = 'Brak danych — wygeneruj pierwszą wycenę, aby zobaczyć statystyki';
        statsSection.appendChild(empty);
        return;
      }

      const aktywnyOkres = STATS_OKRESY.some(o => o.id === okresArg)
        ? okresArg
        : wczytajOkresStat();
      zapiszOkresStat(aktywnyOkres);

      const okresBar = document.createElement('div');
      okresBar.className = 'stats-okres-bar';
      okresBar.setAttribute('role', 'tablist');
      okresBar.setAttribute('aria-label', 'Filtr okresu statystyk');
      STATS_OKRESY.forEach(o => {
        const chip = document.createElement('button');
        chip.type = 'button';
        chip.className = 'chip' + (o.id === aktywnyOkres ? ' is-active' : '');
        chip.textContent = o.label;
        chip.setAttribute('aria-pressed', o.id === aktywnyOkres ? 'true' : 'false');
        chip.dataset.okres = o.id;
        chip.addEventListener('click', () => {
          if (o.id === aktywnyOkres) return;
          renderStatystyki(o.id);
        });
        okresBar.appendChild(chip);
      });
      statsSection.appendChild(okresBar);

      const lista = filtrujHistoriePoOkresie(calaHistoria, aktywnyOkres);

      if (lista.length === 0) {
        const empty = document.createElement('div');
        empty.className = 'stats-empty';
        empty.textContent = 'Brak wycen w wybranym okresie';
        statsSection.appendChild(empty);
        return;
      }

      const teraz = new Date();
      const aktMies = teraz.getMonth();
      const aktRok = teraz.getFullYear();

      const liczbaLacznie = lista.length;
      let liczbaWMiesiacu = 0;
      let sumaLacznie = 0;
      let sumaWMiesiacu = 0;
      let sumaZyskuLacznie = 0;
      let sumaZyskuWMiesiacu = 0;
      let liczbaZeZyskiem = 0;
      let liczbaZeZyskiemWMiesiacu = 0;
      const liczniki = new Map();

      lista.forEach(wpis => {
        const data = new Date(wpis.dataZapisu);
        const wMiesiacu = !Number.isNaN(data.getTime())
          && data.getMonth() === aktMies
          && data.getFullYear() === aktRok;
        const suma = Number(wpis.suma) || 0;
        sumaLacznie += suma;
        if (wMiesiacu) {
          liczbaWMiesiacu += 1;
          sumaWMiesiacu += suma;
        }
        const zysk = (wpis.zysk == null) ? null : Number(wpis.zysk);
        if (Number.isFinite(zysk)) {
          sumaZyskuLacznie += zysk;
          liczbaZeZyskiem += 1;
          if (wMiesiacu) {
            sumaZyskuWMiesiacu += zysk;
            liczbaZeZyskiemWMiesiacu += 1;
          }
        }
        const pozycje = (wpis.payload && Array.isArray(wpis.payload.pozycje))
          ? wpis.payload.pozycje
          : [];
        pozycje.forEach(p => {
          const nazwa = p && typeof p.nazwa === 'string' ? p.nazwa.trim() : '';
          if (!nazwa) return;
          liczniki.set(nazwa, (liczniki.get(nazwa) || 0) + 1);
        });
      });

      const top5 = Array.from(liczniki.entries())
        .sort((a, b) => b[1] - a[1])
        .slice(0, 5);

      const topDzien = policzTopDzienTygodnia(lista);

      const grid = document.createElement('div');
      grid.className = 'stats-grid';

      grid.appendChild(tworzKarteStat(
        'Wygenerowane wyceny',
        String(liczbaLacznie),
        'W tym miesiącu: ' + liczbaWMiesiacu
      ));
      grid.appendChild(tworzKarteStat(
        'Wartość wycen',
        formatujSume(sumaLacznie),
        'W tym miesiącu: ' + formatujSume(sumaWMiesiacu)
      ));

      const zyskCard = document.createElement('div');
      zyskCard.className = 'stats-card';
      const zyskLabel = document.createElement('div');
      zyskLabel.className = 'stats-card-label';
      zyskLabel.textContent = 'Szacowany zysk';
      zyskCard.appendChild(zyskLabel);
      const zyskValue = document.createElement('div');
      zyskValue.className = 'stats-card-value';
      if (liczbaZeZyskiem > 0) {
        zyskValue.textContent = formatujSume(sumaZyskuLacznie);
        if (sumaZyskuLacznie > 0.005) {
          zyskValue.classList.add('stats-trend-positive');
        } else if (sumaZyskuLacznie < -0.005) {
          zyskValue.classList.add('stats-trend-negative');
        }
      } else {
        zyskValue.textContent = '—';
      }
      zyskCard.appendChild(zyskValue);
      const zyskMeta = document.createElement('div');
      zyskMeta.className = 'stats-card-meta';
      if (liczbaZeZyskiem > 0) {
        const wMies = formatujSume(sumaZyskuWMiesiacu);
        const procentPokrycia = Math.round((liczbaZeZyskiem / liczbaLacznie) * 100);
        zyskMeta.textContent = 'W tym miesiącu: ' + wMies
          + ' · z ' + liczbaZeZyskiem + '/' + liczbaLacznie + ' wycen (' + procentPokrycia + '%)';
      } else {
        zyskMeta.textContent = 'Uzupełnij "Twój koszt" w nowych wycenach, aby liczyć zysk';
      }
      zyskCard.appendChild(zyskMeta);
      zyskCard.title = 'Suma (cena dla klienta − Twój koszt) × ilość dla wycen, w których uzupełniłeś koszt własny. Liczone tylko lokalnie, na podstawie historii wycen.';
      grid.appendChild(zyskCard);

      const topDzienValue = topDzien.indeks >= 0 ? NAZWY_DNI[topDzien.indeks] : '—';
      const topDzienMeta = topDzien.indeks >= 0
        ? topDzien.liczba + ' ' + sufiksOferty(topDzien.liczba)
        : 'Brak danych';
      grid.appendChild(tworzKarteStat(
        'Top dzień tygodnia',
        topDzienValue,
        topDzienMeta
      ));

      statsSection.appendChild(grid);

      const chartContainer = document.createElement('div');
      chartContainer.className = 'chart-container';
      const chartTitle = document.createElement('h3');
      chartTitle.className = 'chart-title';
      chartTitle.textContent = 'Wartość wycen (ostatnie 6 miesięcy)';
      chartContainer.appendChild(chartTitle);
      const chartBars = document.createElement('div');
      chartBars.id = 'stat-chart';
      chartContainer.appendChild(chartBars);
      const chartLabels = document.createElement('div');
      chartLabels.id = 'stat-chart-labels';
      chartContainer.appendChild(chartLabels);
      statsSection.appendChild(chartContainer);

      const buckets = agregujMiesiacami(lista, 6);
      renderujWykres(buckets);

      const top5Card = tworzKarteStat('Najczęstsze pozycje', null, null, 'stats-card-full stats-card-standalone');
      if (top5.length > 0) {
        const ul = document.createElement('ul');
        ul.className = 'stats-card-list';
        top5.forEach(([nazwa, n]) => {
          const li = document.createElement('li');
          const nameSpan = document.createElement('span');
          nameSpan.className = 'name';
          nameSpan.textContent = nazwa;
          const countSpan = document.createElement('span');
          countSpan.className = 'count';
          countSpan.textContent = n + 'x';
          li.appendChild(nameSpan);
          li.appendChild(countSpan);
          ul.appendChild(li);
        });
        top5Card.appendChild(ul);
      } else {
        const meta = document.createElement('div');
        meta.className = 'stats-card-meta';
        meta.textContent = 'Brak pozycji do zliczenia';
        top5Card.appendChild(meta);
      }
      statsSection.appendChild(top5Card);

      const klienciCard = renderujKarteTopKlienci(lista, aktywnyOkres);
      statsSection.appendChild(klienciCard);

      const heatmapCard = renderujKarteHeatmapy(calaHistoria);
      statsSection.appendChild(heatmapCard);
    }

    function renderujHistorie() {
      const lista = wczytajHistorie();
      historiaList.innerHTML = '';
      btnHistoriaWyczysc.hidden = lista.length === 0;

      if (lista.length === 0) {
        const li = document.createElement('li');
        li.className = 'historia-empty';
        li.textContent = 'Brak zapisanych wycen';
        historiaList.appendChild(li);
        return;
      }

      lista.forEach(wpis => {
        const li = document.createElement('li');
        li.className = 'historia-item';

        const info = document.createElement('div');
        info.className = 'historia-item-info';

        const row1 = document.createElement('div');
        row1.className = 'historia-item-row';
        const numer = document.createElement('span');
        numer.className = 'historia-item-numer';
        numer.textContent = wpis.numerOferty || 'Bez numeru';
        const suma = document.createElement('span');
        suma.className = 'historia-item-suma';
        suma.textContent = formatujSume(wpis.suma);

        // Badge typu dokumentu (F3)
        const docType = wpis.payload && wpis.payload.typ_dokumentu;
        if (docType) {
          const docBadge = document.createElement('span');
          docBadge.className = 'badge-doc-type';
          if (docType === 'faktura_vat') { docBadge.textContent = 'Faktura VAT'; docBadge.classList.add('is-faktura'); }
          else if (docType === 'faktura_proforma') { docBadge.textContent = 'Pro Forma'; docBadge.classList.add('is-proforma'); }
          row1.appendChild(docBadge);
        }

        row1.appendChild(numer);
        row1.appendChild(suma);

        const row2 = document.createElement('div');
        row2.className = 'historia-item-row historia-item-meta';
        const klient = document.createElement('span');
        klient.className = 'historia-item-klient';
        klient.textContent = wpis.klient || 'Bez nazwy klienta';
        const data = document.createElement('span');
        data.className = 'historia-item-data';
        data.textContent = formatujDateZapisu(wpis.dataZapisu);
        row2.appendChild(klient);
        row2.appendChild(data);

        info.appendChild(row1);
        info.appendChild(row2);

        if (wpis.zysk != null && Number.isFinite(Number(wpis.zysk))) {
          const zyskNum = Number(wpis.zysk);
          const row3 = document.createElement('div');
          row3.className = 'historia-item-row historia-item-meta';
          const zyskInfo = document.createElement('span');
          zyskInfo.textContent = 'Szacowany zysk: ' + formatujSume(zyskNum);
          if (zyskNum > 0.005) zyskInfo.classList.add('stats-trend-positive');
          else if (zyskNum < -0.005) zyskInfo.classList.add('stats-trend-negative');
          row3.appendChild(zyskInfo);
          info.appendChild(row3);
        }

        // Akceptacja wyceny (F2)
        if (wpis.token) {
          const rowAkceptacja = document.createElement('div');
          rowAkceptacja.className = 'historia-item-row historia-item-meta';
          const badgeAkceptacji = document.createElement('span');
          badgeAkceptacji.className = 'badge-akceptacja';
          badgeAkceptacji.dataset.token = wpis.token;
          if (wpis.akceptacja && wpis.akceptacja.accepted) {
            const ts = new Date(wpis.akceptacja.acceptedAt);
            const datStr = ts.toLocaleString('pl-PL', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' });
            const kto = wpis.akceptacja.imie ? ' · ' + wpis.akceptacja.imie : '';
            badgeAkceptacji.textContent = 'Zaakceptowana ' + datStr + kto;
            badgeAkceptacji.classList.add('is-accepted');
          } else {
            badgeAkceptacji.textContent = 'Oczekuje na akceptację';
            badgeAkceptacji.classList.add('is-pending');
            // Sprawdź status na serwerze asynchronicznie
            (async () => {
              try {
                const res = await fetch('/api/accept?token=' + encodeURIComponent(wpis.token));
                if (!res.ok) return;
                const data = await res.json();
                if (data.accepted) {
                  // Zaktualizuj cache w historii
                  const lista = wczytajHistorie();
                  const idx = lista.findIndex(w => w.id === wpis.id);
                  if (idx !== -1) {
                    lista[idx].akceptacja = data;
                    zapiszHistorie(lista);
                  }
                  const ts = new Date(data.acceptedAt);
                  const datStr = ts.toLocaleString('pl-PL', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' });
                  const kto = data.imie ? ' · ' + data.imie : '';
                  badgeAkceptacji.textContent = 'Zaakceptowana ' + datStr + kto;
                  badgeAkceptacji.classList.remove('is-pending');
                  badgeAkceptacji.classList.add('is-accepted');
                }
              } catch (_) {}
            })();
          }
          rowAkceptacja.appendChild(badgeAkceptacji);
          info.appendChild(rowAkceptacja);
        }

        const actions = document.createElement('div');
        actions.className = 'historia-item-actions';

        const btnLoad = document.createElement('button');
        btnLoad.type = 'button';
        btnLoad.className = 'btn-historia-action';
        btnLoad.textContent = 'Wczytaj do formularza';
        btnLoad.addEventListener('click', () => wczytajWpisDoFormularza(wpis));

        const btnPdf = document.createElement('button');
        btnPdf.type = 'button';
        btnPdf.className = 'btn-historia-action';
        btnPdf.textContent = 'Pobierz PDF ponownie';
        btnPdf.addEventListener('click', () => pobierzPdfZHistorii(wpis, btnPdf));

        const btnDel = document.createElement('button');
        btnDel.type = 'button';
        btnDel.className = 'btn-historia-action is-delete';
        btnDel.textContent = 'Usuń';
        btnDel.setAttribute('aria-label', 'Usuń wpis z historii');
        btnDel.addEventListener('click', () => {
          usunZHistorii(wpis.id);
          renderujHistorie();
          renderStatystyki();
        });

        actions.appendChild(btnLoad);
        actions.appendChild(btnPdf);

        // Przycisk "Wystaw fakturę" (F3) — tylko dla ofert, nie dla faktur
        const docTypeWpis = wpis.payload && wpis.payload.typ_dokumentu;
        const isAlreadyInvoice = (docTypeWpis === 'faktura_vat' || docTypeWpis === 'faktura_proforma');
        if (!isAlreadyInvoice) {
          const btnFaktura = document.createElement('button');
          btnFaktura.type = 'button';
          btnFaktura.className = 'btn-historia-action';
          btnFaktura.textContent = 'Wystaw fakturę';
          btnFaktura.addEventListener('click', () => wystawFaktureZHistorii(wpis));
          actions.appendChild(btnFaktura);
        }

        actions.appendChild(btnDel);

        li.appendChild(info);
        li.appendChild(actions);
        historiaList.appendChild(li);
      });
    }

    function wczytajWpisDoFormularza(wpis) {
      if (!wpis || !wpis.payload) return;
      const ok = window.confirm('Wczytanie nadpisze obecny szkic wyceny. Kontynuować?');
      if (!ok) return;

      const p = wpis.payload;
      const klientEl = document.getElementById('klient');
      const numerEl = document.getElementById('numer_oferty');
      const dataEl = document.getElementById('data_waznosci');
      const uwagiEl = document.getElementById('uwagi');

      if (klientEl) klientEl.value = String(p.klient || '');
      if (numerEl) numerEl.value = String(p.numer_oferty || '');
      if (dataEl) dataEl.value = String(p.data_waznosci || '');
      if (uwagiEl) uwagiEl.value = String(p.uwagi || '');

      const koszty = Array.isArray(wpis.koszty) ? wpis.koszty : [];
      tbody.innerHTML = '';
      if (Array.isArray(p.pozycje) && p.pozycje.length > 0) {
        p.pozycje.forEach((poz, idx) => {
          dodajWiersz({ naKoniec: true });
          const tr = tbody.lastElementChild;
          if (!tr) return;
          const inN = tr.querySelector('.in-nazwa');
          const inI = tr.querySelector('.in-ilosc');
          const inC = tr.querySelector('.in-cena');
          const inK = tr.querySelector('.in-koszt');
          const inV = tr.querySelector('.in-vat');
          if (inN && typeof poz.nazwa === 'string') inN.value = poz.nazwa;
          const ilosc = Number(poz && poz.ilosc);
          if (inI && Number.isFinite(ilosc)) inI.value = String(ilosc);
          const cena = Number(poz && poz.cena_jednostkowa);
          if (inC && Number.isFinite(cena)) inC.value = String(cena);
          const koszt = Number(koszty[idx]);
          if (inK && Number.isFinite(koszt)) inK.value = String(koszt);
          if (inV && poz.stawka_vat != null) inV.value = String(poz.stawka_vat);
        });
      } else {
        dodajWiersz();
      }

      // Przywróć typ dokumentu z historii
      if (typeof setActiveDocType === 'function') {
        setActiveDocType(p.typ_dokumentu || '');
      }
      // Przywróć pola faktury
      const numerFakturyEl = document.getElementById('numer_faktury');
      const dataSprzedazyEl = document.getElementById('data_sprzedazy');
      const terminPlatnosciEl = document.getElementById('termin_platnosci');
      if (numerFakturyEl) numerFakturyEl.value = String(p.numer_faktury || '');
      if (dataSprzedazyEl) dataSprzedazyEl.value = String(p.data_sprzedazy || '');
      if (terminPlatnosciEl) terminPlatnosciEl.value = String(p.termin_platnosci || '');

      odswiezStanPresetow();
      aktualizujSzacowanyZysk();
      saveDraft();
      zamknijHistorie();
      pokazKomunikat('Wczytano wycenę z historii.', 'success');
    }

    async function pobierzPdfZHistorii(wpis, btn) {
      if (!wpis || !wpis.payload) return;
      const oryginalnyTekst = btn.textContent;
      btn.disabled = true;
      btn.textContent = 'Pobieranie...';
      try {
        const res = await fetch('/quote', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(wpis.payload),
        });
        if (!res.ok) {
          const tekst = await res.text();
          throw new Error(tekst || `Błąd ${res.status}`);
        }
        const blob = await res.blob();
        const url = URL.createObjectURL(blob);
        const klient = String(wpis.payload.klient || '').replace(/[^a-z0-9-_]+/gi, '_') || 'dokument';
        const a = document.createElement('a');
        a.href = url;
        a.download = `wycena-${klient}.pdf`;
        document.body.appendChild(a);
        a.click();
        a.remove();
        setTimeout(() => URL.revokeObjectURL(url), 0);
      } catch (err) {
        window.alert('Nie udało się pobrać PDF: ' + err.message);
      } finally {
        btn.disabled = false;
        btn.textContent = oryginalnyTekst;
      }
    }

    function otworzHistorie() {
      renderujHistorie();
      historiaModal.hidden = false;
      historiaModal.setAttribute('aria-hidden', 'false');
      document.body.style.overflow = 'hidden';
      setTimeout(() => {
        const pierwszaAkcja = historiaList.querySelector('.btn-historia-action');
        if (pierwszaAkcja) pierwszaAkcja.focus();
        else if (btnZamknijHistoria) btnZamknijHistoria.focus();
      }, 0);
    }

    function zamknijHistorie() {
      historiaModal.hidden = true;
      historiaModal.setAttribute('aria-hidden', 'true');
      document.body.style.overflow = '';
    }

    btnHistoria.addEventListener('click', otworzHistorie);
    const btnHistoriaMobile = document.getElementById('btn-historia-mobile');
    if (btnHistoriaMobile) {
      btnHistoriaMobile.addEventListener('click', () => {
        zamknijAppSettings();
        window.setTimeout(otworzHistorie, 300);
      });
    }
    btnZamknijHistoria.addEventListener('click', zamknijHistorie);
    historiaBackdrop.addEventListener('click', zamknijHistorie);
    btnHistoriaWyczysc.addEventListener('click', () => {
      const ok = window.confirm('Wyczyścić całą historię wycen? Tej operacji nie można cofnąć.');
      if (!ok) return;
      zapiszHistorie([]);
      renderujHistorie();
      renderStatystyki();
    });
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && !historiaModal.hidden) zamknijHistorie();
    });

    const kalkulatorModal = document.getElementById('kalkulator-modal');
    const kalkulatorBackdrop = document.getElementById('kalkulator-modal-backdrop');
    const kalkDlugoscEl = document.getElementById('kalk-dlugosc');
    const kalkSzerokoscEl = document.getElementById('kalk-szerokosc');
    const kalkWynikEl = document.getElementById('kalk-wynik');
    const btnKalkWstaw = document.getElementById('btn-kalk-wstaw');
    const btnKalkZamknij = document.getElementById('btn-kalk-zamknij');

    let aktywnyWierszKalkulatora = null;

    function parsujLiczbeKalk(raw) {
      const s = String(raw == null ? '' : raw).replace(',', '.').trim();
      if (!s) return NaN;
      return parseFloat(s);
    }

    function obliczPowierzchnieKalk() {
      const d = parsujLiczbeKalk(kalkDlugoscEl.value);
      const s = parsujLiczbeKalk(kalkSzerokoscEl.value);
      if (!Number.isFinite(d) || !Number.isFinite(s) || d <= 0 || s <= 0) return NaN;
      return Math.round(d * s * 100) / 100;
    }

    function formatujPowierzchnie(n) {
      return n.toFixed(2).replace('.', ',');
    }

    function odswiezKalkulatorWynik() {
      const wynik = obliczPowierzchnieKalk();
      if (Number.isFinite(wynik) && wynik > 0) {
        kalkWynikEl.textContent = 'Powierzchnia: ' + formatujPowierzchnie(wynik) + ' m²';
        btnKalkWstaw.disabled = false;
      } else {
        kalkWynikEl.textContent = 'Powierzchnia: 0,00 m²';
        btnKalkWstaw.disabled = true;
      }
    }

    function otworzKalkulator(tr) {
      if (!tr || !tbody.contains(tr)) return;
      aktywnyWierszKalkulatora = tr;
      kalkDlugoscEl.value = '';
      kalkSzerokoscEl.value = '';
      odswiezKalkulatorWynik();
      kalkulatorModal.hidden = false;
      kalkulatorModal.setAttribute('aria-hidden', 'false');
      document.body.style.overflow = 'hidden';
      setTimeout(() => kalkDlugoscEl.focus(), 0);
    }

    function zamknijKalkulator() {
      kalkulatorModal.hidden = true;
      kalkulatorModal.setAttribute('aria-hidden', 'true');
      document.body.style.overflow = '';
      aktywnyWierszKalkulatora = null;
    }

    function wstawWynikKalkulatora() {
      const wynik = obliczPowierzchnieKalk();
      if (!Number.isFinite(wynik) || wynik <= 0) return;
      const tr = aktywnyWierszKalkulatora;
      if (!tr || !tbody.contains(tr)) {
        zamknijKalkulator();
        return;
      }
      const inIlosc = tr.querySelector('.in-ilosc');
      if (inIlosc) {
        inIlosc.value = String(wynik);
        inIlosc.dispatchEvent(new Event('input', { bubbles: true }));
      }
      zamknijKalkulator();
      saveDraft();
    }

    kalkDlugoscEl.addEventListener('input', odswiezKalkulatorWynik);
    kalkSzerokoscEl.addEventListener('input', odswiezKalkulatorWynik);
    btnKalkWstaw.addEventListener('click', wstawWynikKalkulatora);
    btnKalkZamknij.addEventListener('click', zamknijKalkulator);
    kalkulatorBackdrop.addEventListener('click', zamknijKalkulator);
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && !kalkulatorModal.hidden) zamknijKalkulator();
    });

    initAppSettingsStatsChips();

    const btnAppWyczyscSzkic = document.getElementById('btn-app-wyczysc-szkic');
    const btnAppWyczyscHistorie = document.getElementById('btn-app-wyczysc-historie');

    if (btnAppWyczyscSzkic) {
      btnAppWyczyscSzkic.addEventListener('click', () => {
        const ok = window.confirm('Wyczyścić bieżącą wycenę? Pola formularza wrócą do wartości domyślnych.');
        if (!ok) return;
        wyczyscFormularz();
      });
    }
    if (btnAppWyczyscHistorie) {
      btnAppWyczyscHistorie.addEventListener('click', () => {
        const ok = window.confirm('Wyczyścić całą historię wycen? Tej operacji nie można cofnąć.');
        if (!ok) return;
        zapiszHistorie([]);
        renderujHistorie();
        renderStatystyki();
      });
    }

    // ── Shareable URL: kodowanie / dekodowanie wyceny ────────────────────────

    const URL_SHARE_MAX_LEN = 4000;

    async function kodujWycenaDoURL(payload) {
      const daneBezLoga = Object.assign({}, payload, { logo_base64: '' });
      const json = JSON.stringify(daneBezLoga);
      try {
        if (typeof CompressionStream !== 'undefined') {
          const bytes = new TextEncoder().encode(json);
          const cs = new CompressionStream('deflate-raw');
          const writer = cs.writable.getWriter();
          writer.write(bytes);
          writer.close();
          const chunks = [];
          const reader = cs.readable.getReader();
          for (;;) {
            const { done, value } = await reader.read();
            if (done) break;
            chunks.push(value);
          }
          let totalLen = 0;
          for (const c of chunks) totalLen += c.length;
          const compressed = new Uint8Array(totalLen);
          let off = 0;
          for (const c of chunks) { compressed.set(c, off); off += c.length; }
          let bin = '';
          for (let i = 0; i < compressed.length; i++) bin += String.fromCharCode(compressed[i]);
          const b64 = btoa(bin).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
          return 'v1c' + b64;
        }
      } catch (_) {}
      // Fallback: raw base64 (brak kompresji)
      try {
        const bytes = new TextEncoder().encode(json);
        let bin = '';
        for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i]);
        const b64 = btoa(bin).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
        return 'v1b' + b64;
      } catch (_) { return null; }
    }

    async function dekodujWycenaZURL(str) {
      if (!str || str.length < 4) throw new Error('Brak danych w linku');
      const prefix = str.slice(0, 3);
      const b64 = str.slice(3).replace(/-/g, '+').replace(/_/g, '/');
      const padded = b64 + '='.repeat((4 - b64.length % 4) % 4);
      const bin = atob(padded);
      const bytes = new Uint8Array(bin.length);
      for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);

      if (prefix === 'v1c') {
        const ds = new DecompressionStream('deflate-raw');
        const writer = ds.writable.getWriter();
        writer.write(bytes);
        writer.close();
        const chunks = [];
        const reader = ds.readable.getReader();
        for (;;) {
          const { done, value } = await reader.read();
          if (done) break;
          chunks.push(value);
        }
        let totalLen = 0;
        for (const c of chunks) totalLen += c.length;
        const out = new Uint8Array(totalLen);
        let off = 0;
        for (const c of chunks) { out.set(c, off); off += c.length; }
        return JSON.parse(new TextDecoder().decode(out));
      } else if (prefix === 'v1b') {
        return JSON.parse(new TextDecoder().decode(bytes));
      }
      throw new Error('Nieznany format linku');
    }

    async function initKlientViewer(wParam) {
      const page = document.querySelector('.page');
      if (page) page.classList.add('is-viewer-mode');
      const viewerEl = document.getElementById('view-klient');
      if (viewerEl) { viewerEl.removeAttribute('hidden'); viewerEl.classList.remove('hidden'); }
      const container = document.getElementById('klient-viewer-container');
      if (!container) return;

      // Wyciągnij token akceptacji z URL
      const tParam = new URLSearchParams(location.search).get('t');
      window._klientViewerToken = tParam || null;

      try {
        const dane = await dekodujWycenaZURL(wParam);
        renderKlientViewer(dane);
      } catch (err) {
        container.innerHTML = '';
        const errDiv = document.createElement('div');
        errDiv.className = 'klient-viewer-error card';
        const h = document.createElement('h2');
        h.className = 'klient-viewer-error-title';
        h.textContent = 'Nieprawidłowy lub niekompletny link';
        const p = document.createElement('p');
        p.className = 'klient-viewer-error-desc';
        p.textContent = 'Poproś sprzedawcę o ponowne przesłanie linku do wyceny.';
        errDiv.appendChild(h);
        errDiv.appendChild(p);
        container.appendChild(errDiv);
      }
    }

    function _viewerFormatPLN(v) {
      return Number(v || 0).toFixed(2).replace('.', ',') + ' zł';
    }

    function renderKlientViewer(dane) {
      const container = document.getElementById('klient-viewer-container');
      if (!container) return;
      container.innerHTML = '';

      function el(tag, cls, text) {
        const e = document.createElement(tag);
        if (cls) e.className = cls;
        if (text != null) e.textContent = text;
        return e;
      }

      // Viewer header (poza kartą) – marka
      const viewerHeader = el('div', 'klient-viewer-brand');
      viewerHeader.appendChild(el('span', 'klient-viewer-brand-name', 'Wycena od'));
      viewerHeader.appendChild(el('strong', 'klient-viewer-brand-firma', dane.nazwa_firmy || ''));
      container.appendChild(viewerHeader);

      const card = el('div', 'card klient-viewer-card');

      // Nagłówek dokumentu
      const titleRow = el('div', 'klient-viewer-title-row');
      const docTypeLabel = (dane.typ_dokumentu === 'faktura_vat')
        ? 'FAKTURA VAT'
        : (dane.typ_dokumentu === 'faktura_proforma')
          ? 'FAKTURA PRO FORMA'
          : 'WYCENA';
      titleRow.appendChild(el('h1', 'klient-viewer-doc-title', docTypeLabel));
      if (dane.numer_oferty) titleRow.appendChild(el('span', 'klient-viewer-doc-num', 'Nr ' + dane.numer_oferty));
      card.appendChild(titleRow);

      // Strony: Sprzedawca / Klient
      const partiesRow = el('div', 'klient-viewer-parties');

      const sellerCol = el('div', 'klient-viewer-party');
      sellerCol.appendChild(el('span', 'klient-viewer-party-label', 'Sprzedawca'));
      sellerCol.appendChild(el('strong', 'klient-viewer-party-name', dane.nazwa_firmy || ''));
      const sellerDetails = [];
      if (dane.nip) sellerDetails.push('NIP: ' + dane.nip);
      if (dane.adres) sellerDetails.push(dane.adres);
      if (dane.miasto) sellerDetails.push(dane.miasto);
      if (sellerDetails.length) sellerCol.appendChild(el('p', 'klient-viewer-party-details', sellerDetails.join(' · ')));
      if (dane.telefon) sellerCol.appendChild(el('p', 'klient-viewer-party-details', 'tel. ' + dane.telefon));
      if (dane.email) sellerCol.appendChild(el('p', 'klient-viewer-party-details', 'e-mail: ' + dane.email));

      const buyerCol = el('div', 'klient-viewer-party');
      buyerCol.appendChild(el('span', 'klient-viewer-party-label', 'Klient'));
      const klientLines = String(dane.klient || '').split('\n');
      klientLines.forEach((line, i) => {
        const trimmed = line.trim();
        if (!trimmed) return;
        if (i === 0) buyerCol.appendChild(el('strong', 'klient-viewer-party-name', trimmed));
        else buyerCol.appendChild(el('p', 'klient-viewer-party-details', trimmed));
      });

      partiesRow.appendChild(sellerCol);
      partiesRow.appendChild(buyerCol);
      card.appendChild(partiesRow);

      // Tabela pozycji
      const tableWrap = el('div', 'klient-viewer-table-wrap');
      const table = el('table', 'klient-viewer-table');
      const thead = document.createElement('thead');
      const headerRow = document.createElement('tr');
      [
        { text: 'Lp.', cls: 'col-lp' },
        { text: 'Nazwa', cls: '' },
        { text: 'Ilość', cls: 'col-num' },
        { text: 'Cena jedn.', cls: 'col-num' },
        { text: 'Wartość', cls: 'col-num' },
      ].forEach(({ text, cls }) => {
        const th = document.createElement('th');
        th.textContent = text;
        if (cls) th.className = cls;
        headerRow.appendChild(th);
      });
      thead.appendChild(headerRow);
      table.appendChild(thead);

      const viewerTbody = document.createElement('tbody');
      let total = 0;
      (dane.pozycje || []).forEach((p, i) => {
        const ilosc = Number(p.ilosc) || 0;
        const cena = Number(p.cena_jednostkowa) || 0;
        const wartosc = ilosc * cena;
        total += wartosc;
        const tr = document.createElement('tr');
        [
          { text: String(i + 1), cls: 'col-lp' },
          { text: String(p.nazwa || ''), cls: '' },
          { text: String(ilosc % 1 === 0 ? ilosc : ilosc.toFixed(2).replace('.', ',')), cls: 'col-num' },
          { text: _viewerFormatPLN(cena), cls: 'col-num' },
          { text: _viewerFormatPLN(wartosc), cls: 'col-num' },
        ].forEach(({ text, cls }) => {
          const td = document.createElement('td');
          td.textContent = text;
          if (cls) td.className = cls;
          tr.appendChild(td);
        });
        viewerTbody.appendChild(tr);
      });
      table.appendChild(viewerTbody);
      tableWrap.appendChild(table);
      card.appendChild(tableWrap);

      // Suma
      const totalRow = el('div', 'klient-viewer-total');
      totalRow.appendChild(el('span', 'klient-viewer-total-label', 'Razem:'));
      totalRow.appendChild(el('span', 'klient-viewer-total-value', _viewerFormatPLN(total)));
      card.appendChild(totalRow);

      // Uwagi
      if (dane.uwagi && dane.uwagi.trim()) {
        const notesWrap = el('div', 'klient-viewer-notes');
        notesWrap.appendChild(el('p', 'klient-viewer-notes-label', 'Uwagi'));
        notesWrap.appendChild(el('p', 'klient-viewer-notes-text', dane.uwagi));
        card.appendChild(notesWrap);
      }

      // Ważność
      if (dane.data_waznosci) {
        card.appendChild(el('p', 'klient-viewer-validity', 'Wycena ważna do: ' + dane.data_waznosci));
      }

      // CTA
      const ctaWrap = el('div', 'klient-viewer-cta');
      const btnPdf = el('button', 'btn-primary', 'Pobierz PDF');
      btnPdf.type = 'button';
      btnPdf.id = 'klient-btn-pobierz-pdf';
      let pdfLoading = false;
      btnPdf.addEventListener('click', async () => {
        if (pdfLoading) return;
        pdfLoading = true;
        btnPdf.disabled = true;
        btnPdf.textContent = 'Generowanie…';
        try {
          const res = await fetch('/quote', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(dane),
          });
          if (!res.ok) throw new Error((await res.text()) || 'Błąd ' + res.status);
          const blob = await res.blob();
          const url = URL.createObjectURL(blob);
          const klientName = String(dane.klient || '').split('\n')[0].replace(/[^a-z0-9-_]+/gi, '_') || 'wycena';
          const a = document.createElement('a');
          a.href = url;
          a.download = 'wycena-' + klientName + '.pdf';
          document.body.appendChild(a);
          a.click();
          a.remove();
          setTimeout(() => URL.revokeObjectURL(url), 5000);
        } catch (err) {
          alert('Nie udało się pobrać PDF: ' + err.message);
        } finally {
          pdfLoading = false;
          btnPdf.disabled = false;
          btnPdf.textContent = 'Pobierz PDF';
        }
      });
      ctaWrap.appendChild(btnPdf);
      ctaWrap.appendChild(el('p', 'klient-viewer-hint', 'Link nie zawiera logo firmy. Pobierz PDF, aby zobaczyć pełną wycenę.'));
      card.appendChild(ctaWrap);

      // Sekcja akceptacji (F2)
      const akceptacjaWrap = el('div', 'klient-viewer-akceptacja');
      akceptacjaWrap.id = 'klient-viewer-akceptacja';
      card.appendChild(akceptacjaWrap);

      container.appendChild(card);
      container.appendChild(createPageSignatureElement('klient-viewer-footer'));

      window._klientViewerDane = dane;

      // Załaduj sekcję akceptacji jeśli token jest w URL
      const token = window._klientViewerToken;
      if (token) {
        renderAkceptacjaSection(akceptacjaWrap, token);
      }
    }

    async function renderAkceptacjaSection(wrap, token) {
      wrap.innerHTML = '';
      const sep = document.createElement('div');
      sep.className = 'klient-viewer-sep';
      wrap.appendChild(sep);

      // Sprawdź aktualny status
      let status = null;
      try {
        const res = await fetch('/api/accept?token=' + encodeURIComponent(token));
        if (res.ok) status = await res.json();
      } catch (_) {}

      function el(tag, cls, text) {
        const e = document.createElement(tag);
        if (cls) e.className = cls;
        if (text != null) e.textContent = text;
        return e;
      }

      if (status && status.accepted) {
        const box = el('div', 'akceptacja-box is-accepted');
        const icon = document.createElement('span');
        icon.className = 'akceptacja-icon';
        icon.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" width="20" height="20"><polyline points="20 6 9 17 4 12"/></svg>';
        box.appendChild(icon);
        const txt = el('div', 'akceptacja-txt');
        const ts = new Date(status.acceptedAt);
        const datStr = ts.toLocaleString('pl-PL', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' });
        const kto = status.imie ? ' przez ' + status.imie : '';
        txt.appendChild(el('strong', '', 'Wycena zaakceptowana'));
        txt.appendChild(el('p', 'akceptacja-meta', datStr + kto));
        box.appendChild(txt);
        wrap.appendChild(box);
        return;
      }

      const box = el('div', 'akceptacja-box');
      box.appendChild(el('p', 'akceptacja-label', 'Akceptuję tę wycenę'));
      const imieInput = document.createElement('input');
      imieInput.type = 'text';
      imieInput.className = 'akceptacja-imie';
      imieInput.placeholder = 'Twoje imię i nazwisko (opcjonalnie)';
      imieInput.maxLength = 60;
      box.appendChild(imieInput);
      const btnAkceptuj = el('button', 'btn-primary btn-akceptuj', 'Akceptuję wycenę');
      btnAkceptuj.type = 'button';
      let akceptujLoading = false;
      btnAkceptuj.addEventListener('click', async () => {
        if (akceptujLoading) return;
        akceptujLoading = true;
        btnAkceptuj.disabled = true;
        btnAkceptuj.textContent = 'Wysyłanie…';
        try {
          const res = await fetch('/api/accept', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ token: token, imie: imieInput.value.trim() }),
          });
          if (!res.ok) throw new Error((await res.text()) || 'Błąd ' + res.status);
          const data = await res.json();
          trackEvent('client_accepted');
          renderAkceptacjaSection(wrap, token);
        } catch (err) {
          btnAkceptuj.disabled = false;
          btnAkceptuj.textContent = 'Akceptuję wycenę';
          akceptujLoading = false;
          alert('Nie udało się zapisać akceptacji: ' + err.message);
        }
      });
      box.appendChild(btnAkceptuj);
      wrap.appendChild(box);
    }

    // ── Faktura: Wystaw fakturę z historii ───────────────────────────────────

    function wystawFaktureZHistorii(wpis) {
      if (!wpis || !wpis.payload) return;
      // Wczytaj wycenę do formularza
      wczytajWpisDoFormularza(wpis);
      // Przełącz na typ "Faktura VAT" i ustaw nowy numer faktury
      if (typeof setActiveDocType === 'function') setActiveDocType('faktura_vat');
      const nrFakturyEl = document.getElementById('numer_faktury');
      if (nrFakturyEl && !nrFakturyEl.value.trim()) {
        nrFakturyEl.value = nastepnyNumerFaktury();
      }
      // Ustaw dzisiejszą datę jako datę sprzedaży
      const dataSprzedazyEl = document.getElementById('data_sprzedazy');
      if (dataSprzedazyEl && !dataSprzedazyEl.value) {
        dataSprzedazyEl.value = new Date().toISOString().slice(0, 10);
      }
      saveDraft();
    }

    // ── Faktura: numeracja ────────────────────────────────────────────────────

    function wczytajNumeracjeFaktury() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY_NUMERACJA_FAKTURY);
        if (!raw) return { ostatniNumer: 0 };
        const dane = JSON.parse(raw);
        const n = Number(dane && dane.ostatniNumer);
        return { ostatniNumer: Number.isFinite(n) && n >= 0 ? Math.floor(n) : 0 };
      } catch (e) { return { ostatniNumer: 0 }; }
    }

    function zapiszNumeracjeFaktury(stan) {
      try { localStorage.setItem(STORAGE_KEY_NUMERACJA_FAKTURY, JSON.stringify(stan)); } catch (e) {}
    }

    function nastepnyNumerFaktury() {
      const { ostatniNumer } = wczytajNumeracjeFaktury();
      const rok = new Date().getFullYear();
      return `FV/${rok}/${String(ostatniNumer + 1).padStart(3, '0')}`;
    }

    function inkrementujNumeracjeFaktury() {
      const stan = wczytajNumeracjeFaktury();
      stan.ostatniNumer = (Number(stan.ostatniNumer) || 0) + 1;
      zapiszNumeracjeFaktury(stan);
    }

    // ── Faktura: przełącznik typu dokumentu ──────────────────────────────────

    function initDocTypeSwitcher() {
      const switcher = document.getElementById('doc-type-switcher');
      if (!switcher) return;
      switcher.querySelectorAll('.chip[data-doc-type]').forEach(btn => {
        btn.addEventListener('click', () => {
          switcher.querySelectorAll('.chip').forEach(b => { b.classList.remove('is-active'); b.setAttribute('aria-pressed', 'false'); });
          btn.classList.add('is-active');
          btn.setAttribute('aria-pressed', 'true');
          onDocTypeChange(btn.dataset.docType || '');
        });
      });
    }

    function getActiveDocType() {
      const btn = document.querySelector('#doc-type-switcher .chip.is-active');
      return btn ? (btn.dataset.docType || '') : '';
    }

    function setActiveDocType(type) {
      const switcher = document.getElementById('doc-type-switcher');
      if (!switcher) return;
      switcher.querySelectorAll('.chip[data-doc-type]').forEach(btn => {
        const match = (btn.dataset.docType || '') === (type || '');
        btn.classList.toggle('is-active', match);
        btn.setAttribute('aria-pressed', match ? 'true' : 'false');
      });
      onDocTypeChange(type || '');
    }

    function onDocTypeChange(type) {
      const invoiceFields = document.getElementById('invoice-fields');
      const vatCols = document.querySelectorAll('.vat-col-hidden');
      const labelNumer = document.getElementById('label-numer-oferty');
      const isInvoice = (type === 'faktura_proforma' || type === 'faktura_vat');
      const isVAT = (type === 'faktura_vat');

      if (invoiceFields) invoiceFields.classList.toggle('hidden', !isInvoice);
      vatCols.forEach(el => el.classList.toggle('vat-col-hidden', !isVAT));

      if (labelNumer) {
        if (type === 'faktura_vat' || type === 'faktura_proforma') labelNumer.textContent = 'Numer wyceny (opcjonalny)';
        else labelNumer.textContent = 'Numer wyceny';
      }

      // Aktualizuj tekst przycisku generuj
      const btnGenerujEl = document.getElementById('btn-generuj');
      if (btnGenerujEl) {
        if (isInvoice) {
          const isMobile = !(window.matchMedia && window.matchMedia('(min-width: 1024px)').matches);
          btnGenerujEl.textContent = isMobile ? 'Podgląd i pobierz' : 'Generuj dokument';
        } else {
          if (typeof aktualizujTekstPrzyciskuGeneruj === 'function') aktualizujTekstPrzyciskuGeneruj();
        }
      }

      // Przycisk XML KSeF i nota — widoczne tylko dla faktura_vat
      const btnXml = document.getElementById('btn-pobierz-xml');
      const ksefHint = document.getElementById('ksef-hint');
      if (btnXml) btnXml.classList.toggle('hidden', !isVAT);
      if (ksefHint) ksefHint.classList.toggle('hidden', !isVAT);

      saveDraft();
    }

    // ── Pobierz XML FA(3) do KSeF ────────────────────────────────────────────
    async function pobierzXmlKSeF() {
      const { gotowy, payload } = budujPayloadZFormularza();
      if (!gotowy) {
        pokazKomunikat('Uzupełnij fakturę (nazwa firmy, klient i przynajmniej jedna pozycja), aby wygenerować XML.', 'error');
        return;
      }
      if (!payload.nip) {
        pokazKomunikat('NIP sprzedawcy jest wymagany do generowania XML KSeF — uzupełnij go w zakładce Moja firma.', 'error');
        return;
      }

      const btn = document.getElementById('btn-pobierz-xml');
      const oryg = btn ? btn.innerHTML : '';
      if (btn) { btn.disabled = true; btn.textContent = 'Generowanie…'; }

      try {
        const res = await fetch('/api/xml', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });
        if (!res.ok) {
          const tekst = await res.text();
          throw new Error(tekst || `Błąd ${res.status}`);
        }
        const blob = await res.blob();

        const numerSlug = String(payload.numer_faktury || payload.numer_oferty || 'faktura')
          .replace(/[^a-zA-Z0-9\-_]/g, '_');
        const klientSlug = String(payload.klient || '').split('\n')[0]
          .replace(/[^a-zA-Z0-9\-_]/g, '_');
        const nazwaPliku = `faktura-${numerSlug}-${klientSlug}.xml`;

        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = nazwaPliku;
        document.body.appendChild(a);
        a.click();
        a.remove();
        setTimeout(() => URL.revokeObjectURL(url), 10000);

        trackEvent('xml_downloaded');
        pokazKomunikat('XML pobrany. Wgraj go na ksef.podatki.gov.pl, aby wysłać fakturę.', 'success');
      } catch (err) {
        pokazKomunikat('Nie udało się wygenerować XML: ' + err.message, 'error');
      } finally {
        if (btn) { btn.disabled = false; btn.innerHTML = oryg; }
      }
    }

    async function skopiujLinkWyceny() {
      const { gotowy, payload } = budujPayloadZFormularza();
      if (!gotowy) {
        pokazKomunikat('Uzupełnij wycenę (nazwa firmy, klient i przynajmniej jedna pozycja), aby wygenerować link.', 'error');
        return;
      }
      const btn = document.getElementById('btn-kopiuj-link');
      const oryg = btn ? btn.textContent : '';
      if (btn) { btn.disabled = true; btn.textContent = 'Generowanie…'; }
      try {
        const str = await kodujWycenaDoURL(payload);
        if (!str) throw new Error('Błąd kodowania');

        // Generuj token akceptacji i przypisz do ostatniego wpisu historii
        const token = (crypto && crypto.randomUUID) ? crypto.randomUUID() : null;
        if (token) {
          const lista = wczytajHistorie();
          // Znajdź wpis pasujący do aktualnej wyceny (ten sam numer i klient)
          const numerAkt = String(payload.numer_oferty || '').trim();
          const klientAkt = String(payload.klient || '').trim();
          const idx = lista.findIndex(w =>
            String(w.numerOferty || '').trim() === numerAkt &&
            String(w.klient || '').trim() === klientAkt
          );
          if (idx !== -1 && !lista[idx].token) {
            lista[idx].token = token;
            zapiszHistorie(lista);
          } else if (idx === -1) {
            // Wycena jeszcze nie w historii — token zostanie dołączony przy najbliższym zapisie
            window._pendingLinkToken = token;
          }
        }

        const tokenParam = token ? '&t=' + token : '';
        const url = location.origin + '/?w=' + str + tokenParam;
        if (url.length > URL_SHARE_MAX_LEN) {
          pokazKomunikat('Wycena jest za duża na link (zbyt wiele pozycji lub długich nazw). Wyślij klientowi PDF.', 'error');
          return;
        }
        try {
          await navigator.clipboard.writeText(url);
          trackEvent('link_copied');
          if (btn) { btn.textContent = 'Skopiowano!'; btn.classList.add('is-copied'); }
          pokazKomunikat('Link skopiowany do schowka! Wyślij go klientowi — otworzy wycenę w przeglądarce.', 'success');
          setTimeout(() => {
            if (btn) { btn.textContent = oryg; btn.classList.remove('is-copied'); btn.disabled = false; }
          }, 2000);
          return;
        } catch (_) {
          window.prompt('Skopiuj link do wyceny (Ctrl+C):', url);
          trackEvent('link_copied');
        }
      } catch (err) {
        pokazKomunikat('Nie udało się wygenerować linku: ' + err.message, 'error');
      } finally {
        if (btn && !btn.classList.contains('is-copied')) {
          btn.disabled = false;
          btn.textContent = oryg;
        }
      }
    }
