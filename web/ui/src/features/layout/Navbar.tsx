import { NavLink } from 'react-router-dom';
import styles from './Navbar.module.css';
import logo from '../../images/logo.svg';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCaretDown } from '@fortawesome/free-solid-svg-icons';
import { useState } from 'react';

function Navbar() {
  const [statusVisible, setStatusVisible] = useState(false);
  const toggleStatus = () => setStatusVisible(!statusVisible);

  return (
    <nav className={styles.navbar}>
      <header>
        <NavLink to="/components">
          <img src={logo} alt="Grafana Agent Logo" title="Grafana Agent" />
        </NavLink>
      </header>
      <ul>
        <li>
          <NavLink to="/dag" className="nav-link">
            DAG
          </NavLink>
        </li>
        <li className={styles.statusLink} onClick={toggleStatus}>
          Status
          <FontAwesomeIcon icon={faCaretDown} className={styles.caret} />
          <ul hidden={!statusVisible}>
            <NavLink to="/status/build-info">
              <li>Runtime and build information</li>
            </NavLink>
            <NavLink to="/status/flags">
              <li>Command-line flags</li>
            </NavLink>
            <NavLink to="/status/config">
              <li>Configuration file</li>
            </NavLink>
          </ul>
        </li>
        <li>
          <a href="https://grafana.com/docs/agent/latest">Help</a>
        </li>
      </ul>
    </nav>
  );
}

export default Navbar;
