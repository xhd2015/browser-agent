import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { useState } from 'react';
import AppGen from './AppGen';
import HarViewer from './HarViewer';
import { useTheme } from './hooks/useTheme';
import './App.css';

function ThemeSwitcher() {
    const { mode, setMode, ThemeModes } = useTheme();
    return (
        <div className="theme-switcher">
            {Object.entries(ThemeModes).map(([label, value]) => (
                <button
                    key={value}
                    className={`theme-btn ${mode === value ? 'active' : ''}`}
                    onClick={() => setMode(value)}
                    title={label}
                >
                    {value === 'light' ? 'Light' : value === 'dark' ? 'Dark' : 'System'}
                </button>
            ))}
        </div>
    );
}

function Home() {
    const [count, setCount] = useState(0);

    return (
        <div style={{ textAlign: 'center', padding: '50px' }}>
            <h1>Welcome to Kool Go-React</h1>
            <p>
                Edit <code>src/App.tsx</code> and save to test HMR
            </p>
            <div className="card">
                <button onClick={() => setCount((count) => count + 1)}>
                    count is {count}
                </button>
            </div>
            <div style={{ marginTop: '20px' }}>
                <Link to="/about" style={{ fontSize: '18px', textDecoration: 'none' }}>
                    Go to About Page
                </Link>
            </div>
        </div>
    );
}

function About() {
    return (
        <div style={{ textAlign: 'center', padding: '50px' }}>
            <h1>About</h1>
            <p>This is a generic about page.</p>
            <Link to="/" style={{ fontSize: '18px', textDecoration: 'none' }}>
                Back to Home
            </Link>
        </div>
    );
}

function App() {
    return (
        <Router>
            <nav className="app-nav">
                <div className="app-nav-links">
                    <Link to="/">Home</Link>
                    <Link to="/about">About</Link>
                    <Link to="/har">HAR Viewer</Link>
                    <Link to="/gen">Generated App</Link>
                </div>
                <ThemeSwitcher />
            </nav>

            <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/about" element={<About />} />
                <Route path="/har" element={<HarViewer />} />
                <Route path="/gen" element={<AppGen />} />
            </Routes>
        </Router>
    );
}

export default App;
