import { NavLink } from 'react-router-dom';
import styles from './Navbar.module.css';
import logo from '../../images/logo.svg';

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
          <a href="https://grafana.com/docs/agent/latest">Help</a>
        </li>
      </ul>
    </nav>
  );
}

export default Navbar;
