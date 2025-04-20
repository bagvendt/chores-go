/**
 * ChoreCard web component
 * Displays a single chore with just an image and gamification elements
 */
class ChoreCard extends HTMLElement {
  /**
   * Initialize the ChoreCard
   */
  constructor() {
    super();
    console.log('ChoreCard constructor');

    /**
     * Timer for tracking long press
     * @type {number|null}
     * @private
     */
    this.pressTimer = null;

    /**
     * Flag to track if press is active
     * @type {boolean}
     * @private
     */
    this.pressStarted = false;

    /**
     * Duration in ms for long press to trigger
     * @type {number}
     * @private
     */
    this.longPressDuration = 1500; // 1500ms for long press

    /**
     * Flag to prevent interactions during animations
     * @type {boolean}
     * @private
     */
    this.animationActive = false;

    /**
     * Y coordinate of touch start
     * @type {number|null}
     * @private
     */
    this.touchStartY = null;

    /**
     * Flag to track if chore was ever completed
     * @type {boolean}
     * @private
     */
    this._wasCompleted = false;
  }

  /**
   * List of observed attributes that will trigger attributeChangedCallback
   */
  static get observedAttributes() {
    return ['image-url', 'title', 'completed', 'points', 'chore-id'];
  }

  /**
   * Component connected callback
   */
  connectedCallback() {
    this.render();
    this.setupEventListeners();
  }

  /**
   * Called when an observed attribute changes
   */
  attributeChangedCallback(name, oldValue, newValue) {
    // Only re-render if a value actually changed
    if (oldValue !== newValue) {
      // If completed changes from true to false, remember it was completed before
      if (name === 'completed' && oldValue === 'true' && newValue === 'false') {
        this._wasCompleted = true;
      }

      this.render();
    }
  }

  /**
   * Set up all event listeners for the component
   */
  setupEventListeners() {
    // Prevent scrolling on iOS
    this.addEventListener(
      'touchstart',
      (e) => {
        e.preventDefault();
        this.startPress(e);
      },
      { passive: false }
    );

    this.addEventListener('mousedown', this.startPress.bind(this));
    this.addEventListener('mouseup', this.endPress.bind(this));
    this.addEventListener('mouseleave', this.cancelPress.bind(this));
    this.addEventListener('touchend', this.endPress.bind(this));
    this.addEventListener('touchcancel', this.cancelPress.bind(this));

    // Prevent context menu from appearing on long press
    this.addEventListener('contextmenu', (e) => {
      e.preventDefault();
      return false;
    });

    // Handle touch move events
    this.addEventListener(
      'touchmove',
      (e) => {
        if ('touches' in e) {
          this.checkTouchMove(e);
        }
      },
      { passive: true }
    );
  }

  /**
   * Get attribute with default value if not present
   */
  getAttr(name, defaultValue = '') {
    return this.hasAttribute(name) ? this.getAttribute(name) : defaultValue;
  }

  /**
   * Render the chore card
   */
  render() {
    const completed = this.getAttr('completed') === 'true';
    const completedClass = completed ? 'completed' : '';
    // Only show status emoji for completed (✅) or previously completed (❌) states
    const statusEmoji = completed ? '✅' : this._wasCompleted ? '❌' : '';
    const points = parseInt(this.getAttr('points', '0'), 10);
    const imageUrl = this.getAttr('image-url');
    const title = this.getAttr('title');

    // Create main CSS if it doesn't exist yet
    if (!document.getElementById('chore-card-styles')) {
      const styleSheet = document.createElement('style');
      styleSheet.id = 'chore-card-styles';
      styleSheet.textContent = `
        chore-card {
          display: block;
          user-select: none;
          -webkit-user-select: none;
          -webkit-touch-callout: none;
          touch-action: none;
        }
        
        .chore-card {
          background-color: white;
          border-radius: 10px;
          box-shadow: 0 4px 8px rgba(59, 47, 38, 0.1);
          transition: transform 0.3s ease, box-shadow 0.3s ease, background-color 0.3s ease;
          cursor: pointer;
          position: relative;
          overflow: hidden;
          width: 100%;
          height: 100%;
          padding: 0;
          aspect-ratio: 1 / 1;
          user-select: none;
          -webkit-user-select: none;
          -webkit-touch-callout: none;
          touch-action: none;
          border: 2px solid transparent;
        }
        
        .chore-card:hover {
          transform: translateY(-5px);
          box-shadow: 0 6px 12px rgba(59, 47, 38, 0.15);
          border-color: #E8B84E;
        }
        
        .chore-card.completed {
          background-color: rgba(163, 177, 128, 0.2);
          border-color: #6A8E59;
        }
        
        .chore-card.pressing {
          transform: scale(0.95);
          box-shadow: 0 2px 4px rgba(59, 47, 38, 0.1);
          animation: crazyShake 0.5s infinite;
          border-color: #C76F3B;
        }
        
        .chore-image {
          width: 100%;
          height: 100%;
          object-fit: cover;
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          pointer-events: none;
          -webkit-user-drag: none;
        }
        
        .status-indicator {
          position: absolute;
          top: 10px;
          right: 10px;
          font-size: 2rem;
          z-index: 10;
          text-shadow: 0 0 5px white, 0 0 5px white;
          filter: drop-shadow(0 0 2px rgba(59, 47, 38, 0.5));
        }
        
        .progress-indicator {
          position: absolute;
          bottom: 0;
          left: 0;
          height: 8px;
          background-color: #6A8E59;
          width: 0%;
          transition: width 0.1s linear;
          z-index: 10;
        }
        
        @keyframes completedAnimation {
          0% { transform: scale(1); }
          50% { transform: scale(1.2); }
          100% { transform: scale(1); }
        }
        
        @keyframes crazyShake {
          0% { transform: scale(0.95) rotate(0deg); }
          10% { transform: scale(0.95) rotate(-10deg) translate(-4px, -2px); }
          20% { transform: scale(0.95) rotate(8deg) translate(7px, 3px); }
          30% { transform: scale(0.92) rotate(-12deg) translate(-7px, 0); }
          40% { transform: scale(0.98) rotate(9deg) translate(4px, -3px); }
          50% { transform: scale(0.94) rotate(-8deg) translate(-2px, 2px); }
          60% { transform: scale(0.97) rotate(10deg) translate(7px, 0); }
          70% { transform: scale(0.92) rotate(-6deg) translate(-7px, 3px); }
          80% { transform: scale(0.97) rotate(5deg) translate(2px, -3px); }
          90% { transform: scale(0.94) rotate(-7deg) translate(-3px, 0); }
          100% { transform: scale(0.95) rotate(0deg); }
        }
        
        .completed-animation {
          animation: completedAnimation 0.5s ease;
        }
        
        .progress-indicator.active {
          background: linear-gradient(90deg, 
            #E8B84E, #C76F3B, #F2B8C6, #6A8E59, #92B6E0, #A3B180, #A65B3A);
          background-size: 700% 100%;
          animation: rainbowProgress 2s linear infinite;
        }
        
        @keyframes rainbowProgress {
          0% { background-position: 0% 50%; }
          100% { background-position: 100% 50%; }
        }

        /* Star Animation Styles */
        .star-container {
          position: absolute;
          top: 0;
          left: 0;
          width: 100%;
          height: 100%;
          pointer-events: none;
          z-index: 100;
          overflow: hidden;
          opacity: 0;
          visibility: hidden;
          transition: opacity 0.3s;
        }
        
        .star-container.active {
          opacity: 1;
          visibility: visible;
        }
        
        .star {
          position: absolute;
          width: 40px;
          height: 40px;
          background-image: url('../../static/img/star-smaller.avif');
          background-size: contain;
          background-repeat: no-repeat;
          background-position: center;
          z-index: 100;
          filter: drop-shadow(0 0 5px #E8B84E);
          opacity: 0;
          transform: scale(0);
        }
        
        @keyframes starBurst {
          0% { 
            opacity: 0;
            transform: scale(0) rotate(0deg); 
          }
          20% { 
            opacity: 1;
            transform: scale(0.5) rotate(90deg); 
          }
          80% { 
            opacity: 1;
            transform: scale(1) rotate(180deg); 
          }
          100% { 
            opacity: 0.2;
            transform: scale(1.2) rotate(270deg); 
          }
        }
        
        @keyframes starFade {
          0% { opacity: 1; }
          100% { opacity: 0; }
        }
        
        .points-indicator {
          position: absolute;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          font-size: 3rem;
          font-weight: bold;
          color: #E8B84E;
          text-shadow: 0 0 10px rgba(59, 47, 38, 0.7);
          z-index: 101;
          opacity: 0;
          pointer-events: none;
        }
        
        @keyframes pointsPopup {
          0% { 
            opacity: 0;
            transform: translate(-50%, -50%) scale(0.5);
          }
          50% { 
            opacity: 1;
            transform: translate(-50%, -50%) scale(1.5);
          }
          80% { 
            opacity: 1;
            transform: translate(-50%, -50%) scale(1.2);
          }
          100% { 
            opacity: 0;
            transform: translate(-50%, -50%) scale(1);
          }
        }
      `;
      document.head.appendChild(styleSheet);
    }

    // Set the inner HTML content
    this.innerHTML = `
      <div class="chore-card ${completedClass}">
        <img class="chore-image" src="${imageUrl}" alt="${title}" draggable="false">
        <div class="status-indicator">${statusEmoji}</div>
        <div class="progress-indicator"></div>
        <div class="star-container"></div>
        <div class="points-indicator">+${points}</div>
      </div>
    `;
  }

  /**
   * Start the press timer for long press
   * @param {Event} e - The event object
   */
  startPress(e) {
    // Prevent default for touch events to avoid scrolling on iOS
    if (e.type === 'touchstart') {
      e.preventDefault();
    }

    if (this.animationActive) return;

    const progressIndicator = this.querySelector('.progress-indicator');
    const card = this.querySelector('.chore-card');

    if (!card) return;

    // Add pressing class for visual feedback
    card.classList.add('pressing');

    if (progressIndicator instanceof HTMLElement) {
      progressIndicator.classList.add('active');
    }

    this.pressStarted = true;

    // Store touch position if it's a touch event
    if (e.type === 'touchstart' && 'touches' in e) {
      if (e.touches[0]) {
        this.touchStartY = e.touches[0].clientY;
      }
    }

    // Start timing for long press
    let progress = 0;
    this.pressTimer = setInterval(() => {
      progress += 100 / (this.longPressDuration / 100); // Increment by percentage per 100ms
      if (progressIndicator instanceof HTMLElement) {
        progressIndicator.style.width = `${progress}%`;
      }

      // Add vibration feedback every 300ms
      if (progress % 20 === 0 && navigator.vibrate) {
        navigator.vibrate(10);
      }

      if (progress >= 100) {
        this.completeLongPress();
      }
    }, 100);
  }

  /**
   * End the press before completion
   */
  endPress() {
    if (!this.pressStarted) return;
    this.cancelPress();
  }

  /**
   * Cancel the current press
   */
  cancelPress() {
    if (this.pressTimer) {
      clearInterval(this.pressTimer);
      this.pressTimer = null;
    }

    const progressIndicator = this.querySelector('.progress-indicator');
    const card = this.querySelector('.chore-card');

    if (progressIndicator instanceof HTMLElement) {
      progressIndicator.style.width = '0%';
      progressIndicator.classList.remove('active');
    }

    if (card) {
      card.classList.remove('pressing');
    }

    this.pressStarted = false;
  }

  /**
   * Check if the touch has moved too far (for cancellation)
   * @param {Event} e - The touch event
   */
  checkTouchMove(e) {
    if (!this.pressStarted || this.touchStartY === null) return;

    // Standard JavaScript check for touch properties
    if (
      e &&
      typeof e === 'object' &&
      'touches' in e &&
      e.touches &&
      e.touches[0] &&
      typeof e.touches[0].clientY === 'number'
    ) {
      const touchY = e.touches[0].clientY;
      const yDiff = Math.abs(touchY - this.touchStartY);

      // Cancel if moved more than 20px
      if (yDiff > 20) {
        this.cancelPress();
      }
    }
  }

  /**
   * Create and animate stars for the reward animation
   * @param {number} numStars - The number of stars to create
   */
  createStarAnimation(numStars) {
    const starContainer = this.querySelector('.star-container');
    const pointsIndicator = this.querySelector('.points-indicator');

    if (!starContainer || !pointsIndicator) return;

    // Clear any existing stars
    starContainer.innerHTML = '';
    starContainer.classList.add('active');

    // Create stars
    const maxStars = Math.min(numStars, 20); // Limit to 20 stars max for performance

    for (let i = 0; i < maxStars; i++) {
      const star = document.createElement('div');
      star.classList.add('star');

      // Random position
      const randomX = Math.random() * 100;
      const randomY = Math.random() * 100;
      star.style.left = `${randomX}%`;
      star.style.top = `${randomY}%`;

      // Random size
      const scale = 0.5 + Math.random() * 1.5;
      star.style.width = `${40 * scale}px`;
      star.style.height = `${40 * scale}px`;

      // Random rotation
      const initialRotation = Math.random() * 360;
      star.style.transform = `rotate(${initialRotation}deg)`;

      // Random delay
      const delay = Math.random() * 0.5;
      star.style.animation = `starBurst 0.6s ${delay}s forwards, starFade 0.5s ${delay + 1.5}s forwards`;

      starContainer.appendChild(star);
    }

    // Animate points indicator
    if (pointsIndicator instanceof HTMLElement) {
      pointsIndicator.style.animation = 'pointsPopup 1.5s forwards';
    }

    // Reset animations after they complete
    setTimeout(() => {
      starContainer.classList.remove('active');
      if (pointsIndicator instanceof HTMLElement) {
        pointsIndicator.style.animation = '';
      }

      setTimeout(() => {
        starContainer.innerHTML = '';
      }, 300);
    }, 2500);
  }

  /**
   * Complete the long press action
   */
  completeLongPress() {
    this.cancelPress();

    // Get previous completion state
    const wasCompletedBefore = this.getAttribute('completed') === 'true';

    // Toggle completion state
    const newCompletedState = !wasCompletedBefore;
    this.setAttribute('completed', newCompletedState.toString());

    // Track if chore was ever completed
    if (wasCompletedBefore) {
      this._wasCompleted = true;
    }

    // Update the appearance
    const card = this.querySelector('.chore-card');
    const statusIndicator = this.querySelector('.status-indicator');

    if (!card || !statusIndicator) return;

    // Prevent interactions during animation
    this.animationActive = true;

    if (newCompletedState) {
      card.classList.add('completed');
      statusIndicator.textContent = '✅';

      // Add completion animation
      card.classList.add('completed-animation');

      // Start star animation if the chore was just completed
      if (!wasCompletedBefore) {
        // Get number of points from attribute
        const points = parseInt(this.getAttribute('points') || '0', 10);
        // Create one star for each point, minimum 5 stars
        this.createStarAnimation(Math.max(5, points));
      }

      // Add stronger vibration feedback for completion
      if (navigator.vibrate) {
        navigator.vibrate([50, 50, 100]);
      }
    } else {
      card.classList.remove('completed');
      // Show X only if it was previously completed
      statusIndicator.textContent = this._wasCompleted ? '❌' : '';

      // Add vibration feedback for un-completion
      if (navigator.vibrate) {
        navigator.vibrate(50);
      }
    }

    // Get chore ID and routine ID
    const choreId = this.getAttribute('chore-id');
    const routineId = this.closest('[data-routine-id]')?.getAttribute('data-routine-id');

    // Update status via API if we have both IDs
    if (choreId && routineId) {
      this.updateChoreCompletionStatus(routineId, choreId, newCompletedState);
    }
    else {
      console.error('Chore ID or Routine ID not found in the DOM.');
    }
    // Remove the animation class after it completes
    setTimeout(() => {
      if (card) {
        card.classList.remove('completed-animation');
      }
      this.animationActive = false;
    }, 500);
  }

  /**
   * Send API request to update chore completion status
   * @param {string} routineId - ID of the routine
   * @param {string} choreId - ID of the chore
   * @param {boolean} completed - New completion status
   */
  updateChoreCompletionStatus(routineId, choreId, completed) {
    fetch(`/api/routine/${routineId}/chore/${choreId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        completed: completed
      })
    })
      .then(response => {
        if (!response.ok) {
          console.error('Failed to update chore status:', response.statusText);
        }
      })
      .catch(error => {
        console.error('Error updating chore status:', error);
      });
  }
}

// Register the custom element
customElements.define('chore-card', ChoreCard);