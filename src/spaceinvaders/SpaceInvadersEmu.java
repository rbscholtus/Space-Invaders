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

import java.io.BufferedInputStream;
import java.util.Scanner;

/**
 *
 * @author Barend Scholtus
 */
public class SpaceInvadersEmu implements I8080Context {

    private int[] memory;
    private I8080 cpu;

    private String[] ROMfiles = new String[] {
        "roms/invaders.h",
        "roms/invaders.g",
        "roms/invaders.f",
        "roms/invaders.e"
    };

    public SpaceInvadersEmu() {
        memory = new int[0x4000];
        readROMs();
        cpu = new I8080(this);
        cpu.reset();
    }

    public void readROMs() {
        byte[] bytes = new byte[0x800];
        BufferedInputStream bis;
        int memAddr = 0;
        
        for (String filename: ROMfiles) {
            bis = null;
            try {
                bis = new BufferedInputStream(
                        getClass().getResourceAsStream(filename));
                int ret = bis.read(bytes);
                if (ret != bytes.length) {
                    System.err.println(filename + " is not valid.");
                    continue;
                }
                for (byte b: bytes) {
                    memory[memAddr++] = b & 0xFF;
                }
            } catch (Exception e) {
                System.err.println("Can't open or read ROM: " + filename);
                e.printStackTrace();
            } finally {
                if (bis != null) {
                    try {
                        bis.close();
                    } catch (Exception ex) {}
                }
            }
        }

        assert memAddr == ROMfiles.length * 0x800 : "Loading failed.";
    }

    public int read(int addr) {
        assert addr == (addr & 0x3FFF) : "Invalid address for read: " + addr;

        return memory[addr];
    }

    public int read2(int addr) {
        assert addr == (addr & 0x3FFF) : "Invalid address for read: " + addr;

        return memory[addr] | (memory[addr+1] << 8);
    }

    public void write(int addr, int data) {
        assert addr == (addr & 0x3FFF) : "Invalid address for write: " + addr;
        assert data == (data & 0xFF) : "Invalid data in write: " + data;

        memory[addr] = data;
    }

    public I8080 getCpu() {
        return cpu;
    }

    public int[] getMemory() {
        return memory;
    }

    public void runTextEmu() {
        Scanner s = new Scanner(System.in);

        System.out.println(cpu.debugString());
        for (;;) {
            s.nextLine();
            cpu.instruction();
            System.out.println(cpu.debugString());
        }
    }

    /**
     * @param args the command line arguments
     */
    public static void main(String[] args) {
        new SpaceInvadersEmu().runTextEmu();
    }
}