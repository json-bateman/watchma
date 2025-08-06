const PORT = 58008;

/**
   * @typedef {Object} Message
   * @property {string} type
   * @property {string} userID
   * @property {any} data
   */
class GameWebSocket {
  constructor() {
    this.ws = null;
    this.userID = this.getUserID();
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
  }

  getUserID() {
    let userID = localStorage.getItem('userID');
    if (!userID) {
      userID = 'user_' +
        Math.random().toString(36).substring(2, 11);
      localStorage.setItem('userID', userID);
    }
    return userID;
  }

  connect() {
    this.ws = new WebSocket(`ws://localhost:${PORT}/ws`);
    this.ws.onopen = () => this.onOpen();
    this.ws.onmessage = (event) => this.onMessage(event);
    this.ws.onclose = () => this.onClose();
    this.ws.onerror = (error) => this.onError(error);
  }

  // Override these methods in your pages
  onOpen() { console.log('Connected'); }
  onMessage(event) { console.log('Message:', event.data); }
  onClose() { this.reconnect(); }
  onError(error) { console.error('WebSocket error:', error); }

  send(/** @type {Message} */data) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  sendText(/** @type {string} */msg) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(msg);
    }
  }

  reconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      setTimeout(() => {
        this.reconnectAttempts++;
        this.connect();
      }, 1000 * this.reconnectAttempts);
    }
  }
}

function sendMsg(
  /** @type {string} */ msg,
  /** @type {boolean} */ raw = false,
  /** @type {string} */ type = "chat",
) {
  if (raw) {
    ws.sendText(msg)
    return
  }
  ws.send({
    userID: ws.userID,
    type,
    msg,
  })
}
