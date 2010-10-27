/*
 *  Copyright 2010 Barend Scholtus
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
package spaceinvaders.gui;

import javax.swing.*;
import spaceinvaders.*;

/**
 *
 * @author Barend Scholtus
 */
public class App extends JFrame {

    public App() {
        super("Intel 8080 with Space Invaders ROMs in a window");
        I8080Context emu = new SpaceInvadersEmu();
        I8080Panel panel = new I8080Panel(emu);
        setContentPane(panel);
        pack();
    }

    public static void main(String[] args) {
        App app = new App();
        app.setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);
        app.setLocationRelativeTo(null);
        app.setVisible(true);
    }
}
