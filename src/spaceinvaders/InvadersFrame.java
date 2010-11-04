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

import i8080.*;
import java.awt.*;
import java.awt.event.*;
import java.awt.image.BufferedImage;
import java.awt.image.DataBufferInt;
import java.awt.image.WritableRaster;
import java.io.BufferedInputStream;

/**
 *
 * @author Barend Scholtus
 */
public class InvadersFrame extends AnimationFrame implements I8080Context {

    public static final int MEMORY_SIZ = 0x4000;
    private int[] memory;
    private I8080 cpu;
    private int cycles;
    public static final int cyclesPerFrame = 2000000 / 60;
    public static final int halfCyclesPerFrame = cyclesPerFrame / 2;
    private String[] ROMfiles = new String[]{
        "roms/invaders.h",
        "roms/invaders.g",
        "roms/invaders.f",
        "roms/invaders.e"
    }; //file, size, crc, offset
    private BufferedImage canvas;
//    private WritableRaster raster;
    private int[] pixels;
    public static final int SI_WIDTH  = 224;
    public static final int SI_HEIGHT = 256;
    public static final int SI_WHITE  = 0xffffff;
    public static final int SI_RED    = 0xff0000;
    public static final int SI_GREEN  = 0x00ff00;
    public static final int SI_BLACK  = 0x000000;

    public static void main(String[] args) {
        final InvadersFrame f = new InvadersFrame();
        f.setVisible(true);
        f.start();
    }

    public InvadersFrame() {
        super(60);
        setTitle("Space Invaders emu");
        addWindowListener(new WindowAdapter() {

            @Override
            public void windowClosing(WindowEvent e) {
                stop();
                System.exit(0);
            }

            @Override
            public void windowOpened(WindowEvent e) {
                Insets ins = getInsets();
                setSize(SI_WIDTH + ins.left + ins.right, SI_HEIGHT + ins.top + ins.bottom);
            }
        });
        setDisplayStats(true);
//        setFullScreen(true);

        // make the frame faster
        setLayout(null);
        removeAll();
        setIgnoreRepaint(true);

        // buffer
        canvas = new BufferedImage(SI_WIDTH, SI_HEIGHT, BufferedImage.TYPE_INT_RGB);
        pixels = ((DataBufferInt) canvas.getRaster().getDataBuffer()).getData();
//        canvas = getGraphicsConfiguration().createCompatibleImage();
//        raster = canvas.getRaster();

        memory = new int[MEMORY_SIZ];
        cpu = new I8080(this);
        readROMs();
    }

    public int readByte(int addr) {
        return memory[addr];
    }

    public int readWord(int addr) {
        return memory[addr] | (memory[addr + 1] << 8);
    }

    public void writeByte(int addr, int data) {
        if (addr < 0x2400) {
            memory[addr] = data;
        } else if (addr < 0x4000) {
            memory[addr] = data;
            setPixels(addr-0x2400, data);
        } else {
            memory[addr - 0x2000] = data;
        }
    }

    public void writeWord(int addr, int dataLow, int dataHigh) {
        if (addr < 0x4000) {
            memory[addr] = dataLow;
            memory[addr + 1] = dataHigh;
        } else {
            memory[addr - 0x2000] = dataLow;
            memory[addr - 0x1fff] = dataHigh;
        }
    }

    public void out(int dev, int data) {
//        System.out.printf("out dev=%d: %x (%b)\n", dev, data, data);
    }

    public int in(int dev) {
//        System.out.printf("in dev=%d: returning 0", dev);
        return 0;
    }

    private void setPixels(int offset, int data) {
        final int x = offset >>> 5;
        final int y = 255 - ((offset & 0x1f) << 3);
        int buff = y * SI_WIDTH + x;
        pixels[buff] = (data & 0x01) != 0 ? SI_WHITE : SI_BLACK;
        buff -= SI_WIDTH;
        pixels[buff] = (data & 0x02) != 0 ? SI_WHITE : SI_BLACK;
        buff -= SI_WIDTH;
        pixels[buff] = (data & 0x04) != 0 ? SI_WHITE : SI_BLACK;
        buff -= SI_WIDTH;
        pixels[buff] = (data & 0x08) != 0 ? SI_WHITE : SI_BLACK;
        buff -= SI_WIDTH;
        pixels[buff] = (data & 0x10) != 0 ? SI_WHITE : SI_BLACK;
        buff -= SI_WIDTH;
        pixels[buff] = (data & 0x20) != 0 ? SI_WHITE : SI_BLACK;
        buff -= SI_WIDTH;
        pixels[buff] = (data & 0x40) != 0 ? SI_WHITE : SI_BLACK;
        buff -= SI_WIDTH;
        pixels[buff] = (data & 0x80) != 0 ? SI_WHITE : SI_BLACK;
    }

    @Override
    public void update() {
//        cycles += cyclesPerFrame;
//        // screen drawing
//        cycles -= cpu.instructions(cycles - halfCyclesPerFrame);
//        // start of vblank
//        cycles -= cpu.interrupt(0xcf);
//        cycles -= cpu.instructions(cycles);
//        // end of vblank
//        cycles -= cpu.interrupt(0xd7);
        // screen drawing
        cpu.instructions(halfCyclesPerFrame);
        // start of vblank
        cpu.interrupt(0xcf);
        cpu.instructions(halfCyclesPerFrame);
        // end of vblank
        cpu.interrupt(0xd7);
//        System.out.println(""+cycles);
    }

    @Override
    public void render(Graphics g) {
        final int offsetX = getInsets().left;
        final int offsetY = getInsets().top;
        g.drawImage(canvas, offsetX, offsetY, null);
//        int x = offsetX, y = offsetY;
//        for (int i = 0x2400; i < 0x4000; i++) {
//            g.setColor((memory[i] & 0x80) == 0 ? Color.black : Color.white);
//            g.drawLine(x, y, x, y++);
//            g.setColor((memory[i] & 0x40) == 0 ? Color.black : Color.white);
//            g.drawLine(x, y, x, y++);
//            g.setColor((memory[i] & 0x20) == 0 ? Color.black : Color.white);
//            g.drawLine(x, y, x, y++);
//            g.setColor((memory[i] & 0x10) == 0 ? Color.black : Color.white);
//            g.drawLine(x, y, x, y++);
//            g.setColor((memory[i] & 0x08) == 0 ? Color.black : Color.white);
//            g.drawLine(x, y, x, y++);
//            g.setColor((memory[i] & 0x04) == 0 ? Color.black : Color.white);
//            g.drawLine(x, y, x, y++);
//            g.setColor((memory[i] & 0x02) == 0 ? Color.black : Color.white);
//            g.drawLine(x, y, x, y++);
//            g.setColor((memory[i] & 0x01) == 0 ? Color.black : Color.white);
//            g.drawLine(x, y, x, y++);
//            if (y >= 256 + offsetY) {
//                y = offsetY;
//                x++;
//            }
//        }
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
//                e.printStackTrace();
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
