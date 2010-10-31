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
package i8080.gui;

import i8080.*;
import java.io.BufferedInputStream;
import javax.swing.JFrame;

/**
 *
 * @author Barend Scholtus
 */
public class I8080Frame extends JFrame implements I8080Context {

    public static final int RAM_SIZ = 0x2000;
    public static final int MEMORY_SIZ = 0x4000;
    private int[] memory;
    private I8080 cpu;
    private String[] ROMfiles = new String[]{
        "../../spaceinvaders/roms/invaders.h",
        "../../spaceinvaders/roms/invaders.g",
        "../../spaceinvaders/roms/invaders.f",
        "../../spaceinvaders/roms/invaders.e"
    };

    public I8080Frame() {
        memory = new int[MEMORY_SIZ];
        readROMs();
        cpu = new I8080(this);
        cpu.reset();

        setTitle("Intel 8080 with Space Invaders ROMs in a window");
        I8080Panel panel = new I8080Panel(cpu, memory);
        setContentPane(panel);
        pack();
    }

    public static void main(String[] args) {
        I8080Frame app = new I8080Frame();
        app.setDefaultCloseOperation(JFrame.EXIT_ON_CLOSE);
        app.setLocationRelativeTo(null);
        app.setVisible(true);
    }

    private void readROMs() {
        byte[] bytes = new byte[0x800];
        BufferedInputStream bis;
        int memAddr = 0;

        for (String filename : ROMfiles) {
            bis = null;
            try {
                bis = new BufferedInputStream(
                        getClass().getResourceAsStream(filename));
                int ret = bis.read(bytes);
                if (ret != bytes.length) {
                    System.err.println(filename + " is not valid.");
                    continue;
                }
                for (byte b : bytes) {
                    memory[memAddr++] = b & 0xFF;
                }
            } catch (Exception e) {
                System.err.println("Can't open or read ROM: " + filename);
                e.printStackTrace();
            } finally {
                if (bis != null) {
                    try {
                        bis.close();
                    } catch (Exception ex) {
                    }
                }
            }
        }

        assert memAddr == ROMfiles.length * 0x800 : "Loading failed.";
    }

    public int read(int addr) {
        return addr < MEMORY_SIZ ? memory[addr] : 0;
    }

    public int read2(int addr) {
        return addr + 1 < MEMORY_SIZ ? memory[addr] | (memory[addr + 1] << 8) : 0;
    }

    public void write(int addr, int data) {
        if (addr >= RAM_SIZ && addr < MEMORY_SIZ) {
            memory[addr] = data;
        }
    }

    public void out(int dev, int data) {
        System.out.printf("out dev=%d: %x (%b)\n", dev, data, data);
    }

    public int in(int dev) {
        System.out.printf("in dev=%d: returning 0", dev);
        return 0;
    }

    public I8080 getCpu() {
        return cpu;
    }

    public int[] getMemory() {
        return memory;
    }
}
