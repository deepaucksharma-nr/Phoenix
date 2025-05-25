// Documentation Feedback System for Phoenix Platform

(function() {
  'use strict';

  // Configuration
  const FEEDBACK_API_URL = 'https://api.phoenix-platform.io/v1/docs/feedback';
  const GITHUB_ISSUE_URL = 'https://github.com/phoenix-platform/phoenix/issues/new';
  
  // Feedback widget HTML
  const feedbackHTML = `
    <div id="feedback-widget" class="feedback-widget">
      <div class="feedback-trigger">
        <span class="feedback-icon">💬</span>
        <span class="feedback-text">Feedback</span>
      </div>
      <div class="feedback-panel">
        <div class="feedback-header">
          <h3>Help us improve this page</h3>
          <button class="feedback-close">×</button>
        </div>
        <div class="feedback-content">
          <div class="feedback-question">
            Was this page helpful?
          </div>
          <div class="feedback-buttons">
            <button class="feedback-btn feedback-yes" data-helpful="yes">
              <span class="emoji">👍</span> Yes
            </button>
            <button class="feedback-btn feedback-no" data-helpful="no">
              <span class="emoji">👎</span> No
            </button>
          </div>
          <div class="feedback-form" style="display: none;">
            <textarea 
              class="feedback-textarea" 
              placeholder="Tell us more (optional)..."
              maxlength="500"
            ></textarea>
            <div class="feedback-actions">
              <button class="feedback-submit">Submit</button>
              <a href="${GITHUB_ISSUE_URL}" target="_blank" class="feedback-github">
                Open GitHub Issue
              </a>
            </div>
          </div>
          <div class="feedback-thanks" style="display: none;">
            <span class="emoji">🙏</span>
            <p>Thank you for your feedback!</p>
          </div>
        </div>
      </div>
    </div>
  `;

  // Inject CSS
  const style = document.createElement('style');
  style.textContent = `
    .feedback-widget {
      position: fixed;
      bottom: 20px;
      right: 20px;
      z-index: 1000;
      font-family: var(--md-text-font-family);
    }

    .feedback-trigger {
      background: var(--md-primary-fg-color);
      color: white;
      padding: 12px 20px;
      border-radius: 30px;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 8px;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      transition: all 0.3s ease;
    }

    .feedback-trigger:hover {
      transform: translateY(-2px);
      box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
    }

    .feedback-panel {
      position: absolute;
      bottom: 70px;
      right: 0;
      width: 320px;
      background: var(--md-default-bg-color);
      border: 1px solid var(--md-default-fg-color--lightest);
      border-radius: 8px;
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
      display: none;
      animation: slideUp 0.3s ease;
    }

    .feedback-panel.active {
      display: block;
    }

    @keyframes slideUp {
      from {
        opacity: 0;
        transform: translateY(10px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }

    .feedback-header {
      padding: 16px;
      border-bottom: 1px solid var(--md-default-fg-color--lightest);
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .feedback-header h3 {
      margin: 0;
      font-size: 16px;
      color: var(--md-default-fg-color);
    }

    .feedback-close {
      background: none;
      border: none;
      font-size: 24px;
      cursor: pointer;
      color: var(--md-default-fg-color--light);
      padding: 0;
      width: 30px;
      height: 30px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 4px;
      transition: all 0.2s ease;
    }

    .feedback-close:hover {
      background: var(--md-default-fg-color--lightest);
    }

    .feedback-content {
      padding: 20px;
    }

    .feedback-question {
      font-size: 14px;
      color: var(--md-default-fg-color);
      margin-bottom: 16px;
      text-align: center;
    }

    .feedback-buttons {
      display: flex;
      gap: 12px;
      justify-content: center;
      margin-bottom: 16px;
    }

    .feedback-btn {
      flex: 1;
      padding: 12px 20px;
      border: 1px solid var(--md-default-fg-color--lightest);
      background: var(--md-default-bg-color);
      border-radius: 6px;
      cursor: pointer;
      font-size: 14px;
      transition: all 0.2s ease;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      color: var(--md-default-fg-color);
    }

    .feedback-btn:hover {
      border-color: var(--md-primary-fg-color);
      background: var(--md-primary-fg-color--transparent);
    }

    .feedback-btn.selected {
      background: var(--md-primary-fg-color);
      color: white;
      border-color: var(--md-primary-fg-color);
    }

    .emoji {
      font-size: 18px;
    }

    .feedback-form {
      margin-top: 16px;
    }

    .feedback-textarea {
      width: 100%;
      min-height: 80px;
      padding: 12px;
      border: 1px solid var(--md-default-fg-color--lightest);
      border-radius: 6px;
      font-size: 14px;
      resize: vertical;
      background: var(--md-code-bg-color);
      color: var(--md-default-fg-color);
      font-family: var(--md-text-font-family);
    }

    .feedback-textarea:focus {
      outline: none;
      border-color: var(--md-primary-fg-color);
    }

    .feedback-actions {
      display: flex;
      gap: 12px;
      margin-top: 12px;
    }

    .feedback-submit {
      flex: 1;
      padding: 10px 20px;
      background: var(--md-primary-fg-color);
      color: white;
      border: none;
      border-radius: 6px;
      cursor: pointer;
      font-size: 14px;
      transition: all 0.2s ease;
    }

    .feedback-submit:hover {
      opacity: 0.9;
    }

    .feedback-github {
      flex: 1;
      padding: 10px 20px;
      border: 1px solid var(--md-default-fg-color--lightest);
      border-radius: 6px;
      text-align: center;
      text-decoration: none;
      color: var(--md-default-fg-color);
      font-size: 14px;
      transition: all 0.2s ease;
    }

    .feedback-github:hover {
      border-color: var(--md-primary-fg-color);
      color: var(--md-primary-fg-color);
    }

    .feedback-thanks {
      text-align: center;
      padding: 20px;
    }

    .feedback-thanks .emoji {
      font-size: 48px;
      display: block;
      margin-bottom: 12px;
    }

    .feedback-thanks p {
      margin: 0;
      color: var(--md-default-fg-color);
    }

    @media (max-width: 768px) {
      .feedback-widget {
        bottom: 10px;
        right: 10px;
      }

      .feedback-panel {
        width: calc(100vw - 20px);
        right: -10px;
      }
    }
  `;
  document.head.appendChild(style);

  // Initialize feedback widget
  function initFeedback() {
    // Add feedback widget to page
    const container = document.createElement('div');
    container.innerHTML = feedbackHTML;
    document.body.appendChild(container.firstElementChild);

    // Get elements
    const widget = document.getElementById('feedback-widget');
    const trigger = widget.querySelector('.feedback-trigger');
    const panel = widget.querySelector('.feedback-panel');
    const closeBtn = widget.querySelector('.feedback-close');
    const feedbackBtns = widget.querySelectorAll('.feedback-btn');
    const form = widget.querySelector('.feedback-form');
    const textarea = widget.querySelector('.feedback-textarea');
    const submitBtn = widget.querySelector('.feedback-submit');
    const thanksMsg = widget.querySelector('.feedback-thanks');

    let feedbackData = {
      page: window.location.pathname,
      helpful: null,
      comment: '',
      timestamp: new Date().toISOString()
    };

    // Toggle panel
    trigger.addEventListener('click', () => {
      panel.classList.toggle('active');
    });

    closeBtn.addEventListener('click', () => {
      panel.classList.remove('active');
    });

    // Handle feedback buttons
    feedbackBtns.forEach(btn => {
      btn.addEventListener('click', (e) => {
        feedbackBtns.forEach(b => b.classList.remove('selected'));
        btn.classList.add('selected');
        feedbackData.helpful = btn.dataset.helpful;
        form.style.display = 'block';
      });
    });

    // Handle submit
    submitBtn.addEventListener('click', async () => {
      feedbackData.comment = textarea.value;
      
      try {
        // Send feedback to API
        await sendFeedback(feedbackData);
        
        // Show thanks message
        form.style.display = 'none';
        widget.querySelector('.feedback-buttons').style.display = 'none';
        widget.querySelector('.feedback-question').style.display = 'none';
        thanksMsg.style.display = 'block';
        
        // Close after 3 seconds
        setTimeout(() => {
          panel.classList.remove('active');
          resetForm();
        }, 3000);
      } catch (error) {
        console.error('Failed to send feedback:', error);
        alert('Failed to send feedback. Please try again or open a GitHub issue.');
      }
    });

    // Reset form
    function resetForm() {
      feedbackBtns.forEach(b => b.classList.remove('selected'));
      form.style.display = 'none';
      thanksMsg.style.display = 'none';
      widget.querySelector('.feedback-buttons').style.display = 'flex';
      widget.querySelector('.feedback-question').style.display = 'block';
      textarea.value = '';
      feedbackData.helpful = null;
      feedbackData.comment = '';
    }

    // Click outside to close
    document.addEventListener('click', (e) => {
      if (!widget.contains(e.target)) {
        panel.classList.remove('active');
      }
    });
  }

  // Send feedback to API
  async function sendFeedback(data) {
    // In production, this would send to your API
    // For now, we'll simulate it
    console.log('Feedback data:', data);
    
    // Track with analytics if available
    if (typeof gtag !== 'undefined') {
      gtag('event', 'feedback', {
        'event_category': 'documentation',
        'event_label': data.page,
        'value': data.helpful === 'yes' ? 1 : 0
      });
    }
    
    // Simulate API call
    return new Promise((resolve) => {
      setTimeout(resolve, 500);
    });
  }

  // Initialize when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initFeedback);
  } else {
    initFeedback();
  }
})();