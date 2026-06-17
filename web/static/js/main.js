(function() {
    if (!window.DISCIPLINE_ID) return;

    var disciplineID = window.DISCIPLINE_ID;
    var feed = document.getElementById('community-feed');
    var emptyState = document.getElementById('empty-state');
    var newCardForm = document.getElementById('new-card-form');
    var ws = null;
    var reconnectTimer = null;

    function connect() {
        var protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
        var url = protocol + '//' + location.host + '/ws/forum/' + disciplineID;

        ws = new WebSocket(url);

        ws.onopen = function() {
            if (reconnectTimer) {
                clearTimeout(reconnectTimer);
                reconnectTimer = null;
            }
        };

        ws.onmessage = function(event) {
            var msg = JSON.parse(event.data);

            switch (msg.type) {
                case 'new_card':
                    addCard(msg.payload);
                    break;
                case 'new_reply':
                    addReply(msg.payload);
                    break;
                case 'delete_card':
                    removeCard(msg.payload.id);
                    break;
            }
        };

        ws.onclose = function() {
            if (!reconnectTimer) {
                reconnectTimer = setTimeout(function() {
                    reconnectTimer = null;
                    connect();
                }, 3000);
            }
        };
    }

    function escapeHtml(text) {
        var div = document.createElement('div');
        div.appendChild(document.createTextNode(text));
        return div.innerHTML;
    }

    function formatDate(isoStr) {
        var d = new Date(isoStr);
        var day = String(d.getDate()).padStart(2, '0');
        var month = String(d.getMonth() + 1).padStart(2, '0');
        var year = d.getFullYear();
        var hours = String(d.getHours()).padStart(2, '0');
        var mins = String(d.getMinutes()).padStart(2, '0');
        return day + '/' + month + '/' + year + ' às ' + hours + ':' + mins;
    }

    function hideEmptyState() {
        if (emptyState) {
            emptyState.style.display = 'none';
        }
    }

    function addCard(card) {
        hideEmptyState();

        var existing = document.querySelector('.community-card[data-card-id="' + card.id + '"]');
        if (existing) return;

        var div = document.createElement('div');
        div.className = 'community-card';
        div.setAttribute('data-card-id', card.id);

        var isOwner = card.is_owner;
        var initial = card.user_name ? card.user_name.charAt(0).toUpperCase() : '?';

        var deleteBtn = '';
        if (isOwner) {
            deleteBtn = '<form class="delete-card-form" data-card-id="' + card.id + '" data-discipline-id="' + disciplineID + '">' +
                '<button type="submit" class="btn btn-danger btn-xs">Excluir</button>' +
                '</form>';
        }

        var title = card.title || '';
        var content = card.content || '';

        div.innerHTML =
            '<div class="card-header">' +
                '<div class="card-author">' +
                    '<span class="community-avatar">' + initial + '</span>' +
                    '<div class="card-author-meta">' +
                        '<span class="community-post-author">' +
                            escapeHtml(card.user_name) +
                            (isOwner ? ' <span class="badge-own">você</span>' : '') +
                        '</span>' +
                        '<span class="community-post-date">' + formatDate(card.created_at) + '</span>' +
                    '</div>' +
                '</div>' +
                deleteBtn +
            '</div>' +
            '<h3 class="card-title">' + escapeHtml(title) + '</h3>' +
            '<div class="community-post-content">' + escapeHtml(content) + '</div>' +
            '<div class="card-replies" data-card-id="' + card.id + '"></div>' +
            '<form class="reply-form" data-card-id="' + card.id + '">' +
                '<input type="hidden" name="card_id" value="' + card.id + '" />' +
                '<div class="community-input-row">' +
                    '<textarea name="content" rows="1" placeholder="Escreva uma resposta..." required></textarea>' +
                    '<button type="submit" class="btn btn-primary btn-sm">Responder</button>' +
                '</div>' +
            '</form>';

        feed.appendChild(div);
    }

    function addReply(reply) {
        var repliesContainer = document.querySelector('.card-replies[data-card-id="' + reply.card_id + '"]');
        if (!repliesContainer) return;

        var existing = repliesContainer.querySelector('.reply[data-reply-id="' + reply.id + '"]');
        if (existing) return;

        var div = document.createElement('div');
        div.className = 'reply';
        div.setAttribute('data-reply-id', reply.id);

        var initial = reply.user_name ? reply.user_name.charAt(0).toUpperCase() : '?';

        div.innerHTML =
            '<div class="reply-header">' +
                '<span class="community-avatar reply-avatar">' + initial + '</span>' +
                '<div class="reply-meta">' +
                    '<span class="reply-author">' + escapeHtml(reply.user_name) + '</span>' +
                    '<span class="community-post-date">' + formatDate(reply.created_at) + '</span>' +
                '</div>' +
            '</div>' +
            '<div class="reply-content">' + escapeHtml(reply.content) + '</div>';

        repliesContainer.appendChild(div);
    }

    function removeCard(cardId) {
        var card = document.querySelector('.community-card[data-card-id="' + cardId + '"]');
        if (card) {
            card.remove();
        }
        var remaining = document.querySelectorAll('.community-card');
        if (remaining.length === 0 && emptyState) {
            emptyState.style.display = 'block';
        }
    }

    function sendMessage(msg) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify(msg));
            return true;
        }
        return false;
    }

    if (newCardForm) {
        newCardForm.addEventListener('submit', function(e) {
            e.preventDefault();

            var titleInput = document.getElementById('card-title-input');
            var contentInput = document.getElementById('card-content-input');
            var title = titleInput.value.trim();
            var content = contentInput.value.trim();

            if (!title || !content) return;

            if (sendMessage({ type: 'new_card', title: title, content: content })) {
                titleInput.value = '';
                contentInput.value = '';
            } else {
                newCardForm.action = '/disciplines/community/post/add';
                newCardForm.method = 'POST';
                newCardForm.submit();
            }
        });
    }

    document.addEventListener('submit', function(e) {
        if (e.target.matches('.reply-form')) {
            e.preventDefault();
            var form = e.target;
            var cardId = parseInt(form.getAttribute('data-card-id'));
            var textarea = form.querySelector('textarea[name="content"]');
            var content = textarea.value.trim();

            if (!content) return;

            if (sendMessage({ type: 'new_reply', card_id: cardId, content: content })) {
                textarea.value = '';
            } else {
                var input = document.createElement('input');
                input.type = 'hidden';
                input.name = 'discipline_id';
                input.value = disciplineID;
                form.appendChild(input);
                form.action = '/forum/reply/add';
                form.method = 'POST';
                form.submit();
            }
        }

        if (e.target.matches('.delete-card-form')) {
            e.preventDefault();
            if (!confirm('Deseja excluir este card?')) return;

            var form = e.target;
            var cardId = parseInt(form.getAttribute('data-card-id'));

            if (sendMessage({ type: 'delete_card', id: cardId })) {
                removeCard(cardId);
            } else {
                var inputId = document.createElement('input');
                inputId.type = 'hidden';
                inputId.name = 'post_id';
                inputId.value = cardId;
                form.appendChild(inputId);

                var inputDis = document.createElement('input');
                inputDis.type = 'hidden';
                inputDis.name = 'discipline_id';
                inputDis.value = disciplineID;
                form.appendChild(inputDis);

                form.action = '/disciplines/community/post/delete';
                form.method = 'POST';
                form.submit();
            }
        }
    });

    connect();
})();
