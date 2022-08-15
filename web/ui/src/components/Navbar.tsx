import { NavLink } from 'react-router-dom';
import styles from './Navbar.module.css';
import logo from '../images/logo.svg';

function Navbar() {
  return (
    <nav className={styles.navbar}>
      <header>
        <NavLink to="/">
          <img src={logo} alt="Grafana Agent Logo" title="Grafana Agent" />
        </NavLink>
      </header>
      <ul>
        <li>
          <NavLink to="/dag" className="nav-link">
            DAG
          </NavLink>
        </li>
        <li>Status</li>
        <li>
          <a href="https://grafana.com/docs/agent/latest">Help</a>
        </li>
      </ul>
    </nav>
  );
}

/*
          <NavLink to="/status/build-info">Runtime and build information</NavLink>
          <NavLink to="/status/flags">Command-line flags</NavLink>
          <NavLink to="/status/config-file">Configuration file</NavLink>

*/

export default Navbar;
