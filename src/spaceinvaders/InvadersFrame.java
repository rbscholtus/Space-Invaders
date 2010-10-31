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
package spaceinvaders;

import i8080.I8080;
import i8080.I8080Context;
import java.awt.Graphics;
import java.io.BufferedInputStream;

/**
 *
 * @author Barend Scholtus
 */
public class InvadersFrame extends AnimationFrame implements I8080Context {

    public static void main(String[] args) {
        final InvadersFrame f = new InvadersFrame();
    }

    public static final int MEMORY_SIZ = 0x4000;
    private int[] memory;
    private I8080 cpu;
    private String[] ROMfiles = new String[]{
        "roms/invaders.h",
        "roms/invaders.g",
        "roms/invaders.f",
        "roms/invaders.e"
    }; //file, size, crc, offset

    public InvadersFrame() {
        super(60);
        setTitle("Space Invaders emu");
        setSize(640, 480);

        memory = new int[MEMORY_SIZ];
        cpu = new I8080(this);
        readROMs();
    }

    public int read(int addr) {
        return memory[addr];
    }

    public int read2(int addr) {
        return memory[addr] | (memory[addr + 1] << 8);
    }

    public void write(int addr, int data) {
        memory[addr] = data;
    }

    public void out(int dev, int data) {
        System.out.printf("out dev=%d: %x (%b)\n", dev, data, data);
    }

    public int in(int dev) {
        System.out.printf("in dev=%d: returning 0", dev);
        return 0;
    }

    @Override
    public void update() {
    }

    @Override
    public void render(Graphics g) {
    }

    public void readROMs() {
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
    }
}
