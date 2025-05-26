/**
 * Interactive Examples for Phoenix Documentation
 * Enables live code editing and execution directly in the documentation
 */

class PhoenixInteractiveExamples {
  constructor() {
    this.examples = document.querySelectorAll('.interactive-example');
    this.initExamples();
  }

  initExamples() {
    this.examples.forEach((example, index) => {
      const id = `example-${index}`;
      example.id = id;
      
      // Create editor and preview containers
      const editorContainer = document.createElement('div');
      editorContainer.className = 'example-editor';
      
      const previewContainer = document.createElement('div');
      previewContainer.className = 'example-preview';
      previewContainer.id = `${id}-preview`;
      
      // Create run button
      const runButton = document.createElement('button');
      runButton.className = 'run-button';
      runButton.textContent = 'Run Example';
      
      // Create reset button
      const resetButton = document.createElement('button');
      resetButton.className = 'reset-button';
      resetButton.textContent = 'Reset';
      
      // Create button container
      const buttonContainer = document.createElement('div');
      buttonContainer.className = 'example-buttons';
      buttonContainer.appendChild(runButton);
      buttonContainer.appendChild(resetButton);
      
      // Extract code and create editor
      const code = example.textContent;
      example.textContent = '';
      
      const editor = this.createEditor(editorContainer, code);
      
      // Append elements
      example.appendChild(editorContainer);
      example.appendChild(buttonContainer);
      example.appendChild(previewContainer);
      
      // Add event listeners
      runButton.addEventListener('click', () => {
        this.runCode(editor.getValue(), previewContainer.id);
      });
      
      resetButton.addEventListener('click', () => {
        editor.setValue(code);
        this.runCode(code, previewContainer.id);
      });
      
      // Initial run
      setTimeout(() => {
        this.runCode(code, previewContainer.id);
      }, 100);
    });
  }
  
  createEditor(container, code) {
    // Use CodeMirror or Monaco Editor here
    // This is a simplified example
    const textarea = document.createElement('textarea');
    textarea.value = code;
    container.appendChild(textarea);
    
    // Initialize CodeMirror (would be added to your dependencies)
    const editor = CodeMirror.fromTextArea(textarea, {
      lineNumbers: true,
      mode: 'javascript',
      theme: 'monokai',
      autoCloseBrackets: true,
      matchBrackets: true,
      tabSize: 2,
    });
    
    return editor;
  }
  
  runCode(code, containerId) {
    const container = document.getElementById(containerId);
    container.innerHTML = '';
    
    try {
      // For React examples
      if (code.includes('React')) {
        this.runReactExample(code, containerId);
        return;
      }
      
      // For REST API examples
      if (code.includes('fetch') || code.includes('axios')) {
        this.runApiExample(code, containerId);
        return;
      }
      
      // For standard JS
      const result = eval(code);
      container.textContent = result !== undefined ? result : 'Code executed successfully';
    } catch (error) {
      container.innerHTML = `<div class="error">${error.message}</div>`;
    }
  }
  
  runReactExample(code, containerId) {
    const container = document.getElementById(containerId);
    
    try {
      const transformedCode = Babel.transform(code, {
        presets: ['react']
      }).code;
      
      eval(transformedCode);
    } catch (error) {
      container.innerHTML = `<div class="error">${error.message}</div>`;
    }
  }
  
  runApiExample(code, containerId) {
    const container = document.getElementById(containerId);
    container.innerHTML = '<div class="loading">Loading response...</div>';
    
    // Create a sandbox to safely execute API calls
    const sandbox = document.createElement('iframe');
    sandbox.style.display = 'none';
    document.body.appendChild(sandbox);
    
    const sandboxContent = `
      <script>
        // Mock API responses
        const mockResponses = {
          '/api/v1/experiments': {
            status: 200,
            data: [
              { id: 1, name: 'Test Experiment', status: 'active' },
              { id: 2, name: 'Performance Test', status: 'completed' }
            ]
          },
          '/api/v1/metrics': {
            status: 200,
            data: {
              cpu: 45.2,
              memory: 1024,
              requests: 1500
            }
          }
        };
        
        // Override fetch
        window.fetch = async (url, options) => {
          console.log('Mock fetching:', url);
          
          // Simulate network delay
          await new Promise(resolve => setTimeout(resolve, 500));
          
          // Find matching mock response
          const mockUrl = Object.keys(mockResponses).find(mock => url.includes(mock));
          
          if (mockUrl) {
            return {
              ok: true,
              status: mockResponses[mockUrl].status,
              json: async () => mockResponses[mockUrl].data
            };
          }
          
          return {
            ok: false,
            status: 404,
            json: async () => ({ error: 'Not found' })
          };
        };
        
        // Execute the code and return result
        try {
          ${code}
          .then(result => {
            window.parent.postMessage({ 
              type: 'apiResult', 
              containerId: '${containerId}',
              result: JSON.stringify(result, null, 2)
            }, '*');
          })
          .catch(error => {
            window.parent.postMessage({ 
              type: 'apiError',
              containerId: '${containerId}',
              error: error.message
            }, '*');
          });
        } catch (error) {
          window.parent.postMessage({ 
            type: 'apiError',
            containerId: '${containerId}',
            error: error.message
          }, '*');
        }
      </script>
    `;
    
    sandbox.srcdoc = sandboxContent;
    
    // Listen for messages from sandbox
    window.addEventListener('message', event => {
      if (event.data.containerId === containerId) {
        if (event.data.type === 'apiResult') {
          container.innerHTML = `<pre class="result">${event.data.result}</pre>`;
        } else if (event.data.type === 'apiError') {
          container.innerHTML = `<div class="error">${event.data.error}</div>`;
        }
        
        // Clean up
        document.body.removeChild(sandbox);
      }
    });
  }
}

// Initialize when the document is ready
document.addEventListener('DOMContentLoaded', () => {
  new PhoenixInteractiveExamples();
});