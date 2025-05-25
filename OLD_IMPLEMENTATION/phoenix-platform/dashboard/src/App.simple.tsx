import React from 'react';

function App() {
  return (
    <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
      <h1>Phoenix Platform Dashboard</h1>
      <p>Process Metrics Optimization Platform</p>
      
      <div style={{ marginTop: '20px' }}>
        <h2>Status</h2>
        <p>Dashboard is running but not fully implemented.</p>
      </div>
      
      <div style={{ marginTop: '20px' }}>
        <h2>Components Status</h2>
        <ul>
          <li>✅ API Service - Implemented (basic)</li>
          <li>✅ Process Simulator - Implemented (basic)</li>
          <li>⚠️ Experiment Controller - Stub only</li>
          <li>⚠️ Config Generator - Stub only</li>
          <li>⚠️ Pipeline Operator - Stub only</li>
          <li>⚠️ LoadSim Operator - Stub only</li>
        </ul>
      </div>
    </div>
  );
}

export default App;