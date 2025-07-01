// Global state
let torrents = [];

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    loadTorrents();
    // Refresh torrents every 2 seconds
    setInterval(loadTorrents, 2000);
});

// Load torrents from API
async function loadTorrents() {
    try {
        const response = await fetch('/api/torrents');
        torrents = await response.json();
        renderTorrents();
    } catch (error) {
        console.error('Failed to load torrents:', error);
    }
}

// Render torrent list
function renderTorrents() {
    const listElement = document.getElementById('torrent-list');
    
    if (torrents.length === 0) {
        listElement.innerHTML = `
            <div class="empty-state">
                <h2>トレントがありません</h2>
                <p>上のボタンからトレントを追加してください</p>
            </div>
        `;
        return;
    }
    
    listElement.innerHTML = torrents.map(torrent => `
        <div class="torrent-item">
            <div class="torrent-info">
                <h3>${escapeHtml(torrent.info.name)}</h3>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${torrent.progress}%"></div>
                </div>
                <div class="torrent-stats">
                    <span class="status status-${torrent.status}">${torrent.status}</span>
                    <span>${formatBytes(torrent.downloaded)} / ${formatBytes(torrent.info.length)}</span>
                    <span>${torrent.progress.toFixed(1)}%</span>
                </div>
            </div>
            <div class="torrent-actions">
                ${torrent.status === 'stopped' ? 
                    `<button class="btn btn-secondary" onclick="startTorrent('${torrent.id}')">開始</button>` :
                    `<button class="btn btn-secondary" onclick="stopTorrent('${torrent.id}')">停止</button>`
                }
                <button class="btn btn-secondary" onclick="removeTorrent('${torrent.id}')">削除</button>
            </div>
        </div>
    `).join('');
}

// Modal functions
function showModal(modalId) {
    document.getElementById(modalId).classList.add('active');
}

function hideModal(modalId) {
    document.getElementById(modalId).classList.remove('active');
}

function showAddTorrentModal() {
    showModal('add-torrent-modal');
}

function showAddMagnetModal() {
    showModal('add-magnet-modal');
}

// Add torrent file
async function addTorrentFile() {
    const fileInput = document.getElementById('torrent-file');
    const file = fileInput.files[0];
    
    if (!file) {
        alert('ファイルを選択してください');
        return;
    }
    
    const formData = new FormData();
    formData.append('torrent', file);
    
    try {
        const response = await fetch('/api/torrents', {
            method: 'POST',
            body: formData
        });
        
        if (response.ok) {
            hideModal('add-torrent-modal');
            fileInput.value = '';
            loadTorrents();
        } else {
            const error = await response.json();
            alert(`エラー: ${error.error}`);
        }
    } catch (error) {
        alert('トレントの追加に失敗しました');
    }
}

// Add magnet link
async function addMagnetLink() {
    const magnetInput = document.getElementById('magnet-link');
    const magnet = magnetInput.value.trim();
    
    if (!magnet) {
        alert('マグネットリンクを入力してください');
        return;
    }
    
    try {
        const response = await fetch('/api/torrents/magnet', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ magnet })
        });
        
        if (response.ok) {
            hideModal('add-magnet-modal');
            magnetInput.value = '';
            loadTorrents();
        } else {
            const error = await response.json();
            alert(`エラー: ${error.error}`);
        }
    } catch (error) {
        alert('マグネットリンクの追加に失敗しました');
    }
}

// Torrent operations
async function startTorrent(id) {
    try {
        const response = await fetch(`/api/torrents/${id}/start`, {
            method: 'POST'
        });
        
        if (response.ok) {
            loadTorrents();
        }
    } catch (error) {
        alert('トレントの開始に失敗しました');
    }
}

async function stopTorrent(id) {
    try {
        const response = await fetch(`/api/torrents/${id}/stop`, {
            method: 'POST'
        });
        
        if (response.ok) {
            loadTorrents();
        }
    } catch (error) {
        alert('トレントの停止に失敗しました');
    }
}

async function removeTorrent(id) {
    if (!confirm('このトレントを削除しますか？')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/torrents/${id}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            loadTorrents();
        }
    } catch (error) {
        alert('トレントの削除に失敗しました');
    }
}

// Utility functions
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}