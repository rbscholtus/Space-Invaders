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
import java.awt.image.*;
import java.io.BufferedInputStream;
import javax.sound.sampled.*;

/**
 *
 * @author Barend Scholtus
 */
public class InvadersFrame extends AnimationFrame implements I8080Context {

    public static final int MEMORY_SIZE = 0x4000;
    public static final int CLOCK_FREQ = 1996800;
    public static final double FRAMES_PER_SEC = 59.541985;
    public static final int CYCLES_PER_FRAME = (int) (CLOCK_FREQ / FRAMES_PER_SEC);
    public static final int HALF_CYCLES_PER_FRAME = CYCLES_PER_FRAME / 2;
    public static final int GAME_WIDTH = 224;
    public static final int GAME_HEIGHT = 256;
    public static final int CLR_WHITE = 0xffffff;
    public static final int CLR_RED = 0xff0000;
    public static final int CLR_GREEN = 0x00ff00;
    public static final int CLR_BLACK = 0x000000;
    public static final int SND_SPACESHIP = 0;
    public static final int SND_SHOT = 1;
    public static final int SND_BASE_HIT = 2;
    public static final int SND_INVADER_HIT = 3;
    public static final int SND_WALK1 = 4;
    public static final int SND_WALK2 = 5;
    public static final int SND_WALK3 = 6;
    public static final int SND_WALK4 = 7;
    public static final int SND_SPACESHIP_HIT = 8; //really?
    public static final int SND_EXT_PLAY = 9; //really?
    public static final String[] ROM_FILES = {
        "roms/invaders.h",
        "roms/invaders.g",
        "roms/invaders.f",
        "roms/invaders.e"
    }; //file, size, crc, offset
    public static final String[] SAMPLE_FILES = {
        "samples/0.wav",
        "samples/1.wav",
        "samples/2.wav",
        "samples/3.wav",
        "samples/4.wav",
        "samples/5.wav",
        "samples/6.wav",
        "samples/7.wav",
        "samples/8.wav",
        "samples/9.wav"
    };
    private I8080 cpu;
    private int[] memory;
    private int in1 = 0x00;
    private int in2 = 0x00;
    private int shiftReg;
    private int shift;
    private BufferedImage canvas;
    private int[] pixels;
    private int scaledWidth = GAME_WIDTH;
    private int scaledHeight = GAME_HEIGHT;
    private int offsetX = 0;
    private Clip[] samples2;
    private boolean[] isPlaying;

    public static void main(String[] args) {
        final InvadersFrame f = new InvadersFrame();
        f.setVisible(true);
        f.start();
    }

    public InvadersFrame() {
        // frame
        super(FRAMES_PER_SEC);
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
                setSize(GAME_WIDTH + ins.left + ins.right, GAME_HEIGHT + ins.top + ins.bottom);
            }
        });
        addKeyListener(new KeyAdapter() {

            @Override
            public void keyPressed(KeyEvent e) {
                if (e.getKeyCode() == KeyEvent.VK_ENTER
                        && (e.getModifiersEx() & KeyEvent.ALT_DOWN_MASK) != 0) {

                    // go fullscreen
                    setFullScreen(true);
                    setESCToExitFullScreen(true);

                    // calculate dimensions
                    scaledHeight = getHeight();
                    scaledWidth = GAME_WIDTH * scaledHeight / canvas.getHeight();
                    offsetX = (getWidth() - scaledWidth) / 2;
                }
            }
        });
        setDisplayStats(true);

        // make the frame faster
        setLayout(null);
        removeAll();
        setIgnoreRepaint(true);
        setBackground(Color.BLACK);

        // buffer
        canvas = new BufferedImage(GAME_WIDTH, GAME_HEIGHT, BufferedImage.TYPE_INT_RGB);
        pixels = ((DataBufferInt) canvas.getRaster().getDataBuffer()).getData();

        // emulator
        memory = new int[MEMORY_SIZE];
        cpu = new I8080(this);
        loadROMs();
        loadSamples();

        // keys
        addKeyListener(new KeyAdapter() {

            @Override
            public void keyPressed(KeyEvent e) {
                keypress(e);
            }

            @Override
            public void keyReleased(KeyEvent e) {
                keyrelease(e);
            }
        });
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
            setPixels(addr - 0x2400, data);
        } else {
            addr = 0x2000 | (addr & 0x3ff);
            memory[addr] = data;
        }
    }

    public void writeWord(int addr, int dataLow, int dataHigh) {
        if (addr < 0x4000) {
            memory[addr] = dataLow;
            memory[addr + 1] = dataHigh;
        } else {
            addr = 0x2000 | (addr & 0x3ff);
            memory[addr] = dataLow;
            memory[addr + 1] = dataHigh;
        }
    }

    public void out(int dev, int data) {
//        System.out.printf("out dev=%d: %x (%b)\n", dev, data, data);
        switch (dev) {
            case 2:
                shift = 8 - data;
                break;
            case 3:
//                System.out.println("OUT3: "+Util.toBinString(data, 8));
                if ((data & 0x01) != 0) {
                    samples2[SND_SPACESHIP].loop(99);
                } else {
                    samples2[SND_SPACESHIP].stop();
                    samples2[SND_SPACESHIP].setFramePosition(0);
                }
                if ((data & 0x02) != 0 && !isPlaying[SND_SHOT]) {
                    isPlaying[SND_SHOT] = true;
                    samples2[SND_SHOT].setFramePosition(0);
                    samples2[SND_SHOT].start();
                }
                if ((data & 0x04) != 0 && !isPlaying[SND_BASE_HIT]) {
                    isPlaying[SND_BASE_HIT] = true;
                    samples2[SND_BASE_HIT].setFramePosition(0);
                    samples2[SND_BASE_HIT].start();
                }
                if ((data & 0x08) != 0 && !isPlaying[SND_INVADER_HIT]) {
                    isPlaying[SND_INVADER_HIT] = true;
                    samples2[SND_INVADER_HIT].setFramePosition(0);
                    samples2[SND_INVADER_HIT].start();
                }
                if ((data & 0x10) != 0 && !isPlaying[SND_EXT_PLAY]) {
                    isPlaying[SND_EXT_PLAY] = true;
                    samples2[SND_EXT_PLAY].setFramePosition(0);
                    samples2[SND_EXT_PLAY].start();
                }
                break;
            case 4:
                shiftReg >>>= 8;
                shiftReg |= (data << 8);
                break;
            case 5:
//                System.out.println("sound OUT5: "+Util.toBinString(data, 8));
                if ((data & 0x01) != 0 && !isPlaying[SND_WALK1]) {
                    samples2[SND_WALK1].setFramePosition(0);
                    samples2[SND_WALK1].start();
                }
                if ((data & 0x02) != 0 && !isPlaying[SND_WALK2]) {
                    samples2[SND_WALK2].setFramePosition(0);
                    samples2[SND_WALK2].start();
                }
                if ((data & 0x04) != 0 && !isPlaying[SND_WALK3]) {
                    samples2[SND_WALK3].setFramePosition(0);
                    samples2[SND_WALK3].start();
                }
                if ((data & 0x08) != 0 && !isPlaying[SND_WALK4]) {
                    samples2[SND_WALK4].setFramePosition(0);
                    samples2[SND_WALK4].start();
                }
                if ((data & 0x10) != 0 && !isPlaying[SND_SPACESHIP_HIT]) {
                    samples2[SND_SPACESHIP_HIT].setFramePosition(0);
                    samples2[SND_SPACESHIP_HIT].start();
                }
                break;
        }
    }

    public int in(int dev) {
        switch (dev) {
            case 1:
                int ret = in1;
                in1 &= 0xfe;
                return ret;
            case 2:
                return in2;
            case 3:
                return (shiftReg >>> shift) & 0xff;
        }
        return 0;
    }

    private void keypress(KeyEvent e) {
        switch (e.getKeyCode()) {
            case KeyEvent.VK_5:
                in1 |= 0x01;
                break;
            case KeyEvent.VK_2:
                in1 |= 0x02;
                break;
            case KeyEvent.VK_1:
                in1 |= 0x04;
                break;
            case KeyEvent.VK_UP:
                in1 |= 0x10;
                break;
            case KeyEvent.VK_LEFT:
                in1 |= 0x20;
                break;
            case KeyEvent.VK_RIGHT:
                in1 |= 0x40;
                break;
            case KeyEvent.VK_NUMPAD8:
                in2 |= 0x10;
                break;
            case KeyEvent.VK_NUMPAD4:
                in2 |= 0x20;
                break;
            case KeyEvent.VK_NUMPAD6:
                in2 |= 0x40;
                break;
        }
    }

    private void keyrelease(KeyEvent e) {
        switch (e.getKeyCode()) {
            case KeyEvent.VK_2:
                in1 &= ~0x02;
                break;
            case KeyEvent.VK_1:
                in1 &= ~0x04;
                break;
            case KeyEvent.VK_UP:
                in1 &= ~0x10;
                break;
            case KeyEvent.VK_LEFT:
                in1 &= ~0x20;
                break;
            case KeyEvent.VK_RIGHT:
                in1 &= ~0x40;
                break;
            case KeyEvent.VK_NUMPAD8:
                in2 &= ~0x10;
                break;
            case KeyEvent.VK_NUMPAD4:
                in2 &= ~0x20;
                break;
            case KeyEvent.VK_NUMPAD6:
                in2 &= ~0x40;
                break;
        }
    }

    private void setPixels(int offset, int data) {
        final int x = offset >>> 5;
        final int y = 248 - ((offset & 0x1f) << 3);
        final int color = (y >= 32 && y < 64) ? CLR_RED
                : (y >= 184 && y < 240) ? CLR_GREEN
                : (y >= 240 && x >= 16 && x < 134) ? CLR_GREEN
                : CLR_WHITE;
        int buff = y * GAME_WIDTH + x;
        pixels[buff] = (data & 0x80) != 0 ? color : CLR_BLACK;
        buff += GAME_WIDTH;
        pixels[buff] = (data & 0x40) != 0 ? color : CLR_BLACK;
        buff += GAME_WIDTH;
        pixels[buff] = (data & 0x20) != 0 ? color : CLR_BLACK;
        buff += GAME_WIDTH;
        pixels[buff] = (data & 0x10) != 0 ? color : CLR_BLACK;
        buff += GAME_WIDTH;
        pixels[buff] = (data & 0x08) != 0 ? color : CLR_BLACK;
        buff += GAME_WIDTH;
        pixels[buff] = (data & 0x04) != 0 ? color : CLR_BLACK;
        buff += GAME_WIDTH;
        pixels[buff] = (data & 0x02) != 0 ? color : CLR_BLACK;
        buff += GAME_WIDTH;
        pixels[buff] = (data & 0x01) != 0 ? color : CLR_BLACK;
    }

    @Override
    public void update() {
        // screen drawing
        cpu.instructions(HALF_CYCLES_PER_FRAME);
        // start of vblank
        cpu.interrupt(0xcf);
        cpu.instructions(HALF_CYCLES_PER_FRAME);
        // end of vblank
        cpu.interrupt(0xd7);
    }

    @Override
    public void render(Graphics g) {
        if (isFullScreen()) {
            g.drawImage(canvas, offsetX, 0, scaledWidth, scaledHeight, null);
        } else {
            g.drawImage(canvas, getInsets().left, getInsets().top, null);
        }
    }

    private void loadROMs() {
        byte[] bytes = new byte[0x800];
        BufferedInputStream bis;
        int memAddr = 0;

        for (String filename : ROM_FILES) {
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

    private void loadSamples() {
        samples2 = new Clip[SAMPLE_FILES.length];
        for (int i = 0; i < SAMPLE_FILES.length; i++) {
            try {
                samples2[i] = AudioSystem.getClip();
                AudioInputStream ais = AudioSystem.getAudioInputStream(
                        getClass().getResource(SAMPLE_FILES[i]));
                samples2[i].open(ais);
                ais.close();
                samples2[i].addLineListener(new LL(i));
            } catch (Exception e) {
                System.err.println(e);
            }

        }
        isPlaying = new boolean[samples2.length];
    }

    class LL implements LineListener {
        private final int i;
        public LL(final int i) {
            this.i = i;
        }

        public void update(LineEvent event) {
            if (event.getType() == LineEvent.Type.STOP) {
                isPlaying[i] = false;
            }
        }
    }

    // For in2
    /** 0x00:3, 0x01:4, 0x02:5, 0x03:6 */
    public int getNumberOfLives() {
        return (in2 & 0x03) + 3;
    }

    public void setNumberOfLives(int numberOfLives) {
        if (numberOfLives < 3 || numberOfLives > 6) {
            throw new IllegalArgumentException("numberOfLives must be 3-6");
        }
        in2 = (in2 & ~0x03) | numberOfLives - 3;
    }

    // For in2
    /** 0x00:1500 0x08:1000 */
    public boolean getBonusLifeAt1000() {
        return (in2 & 0x08) != 0;
    }

    public void setBonusLifeAt1000(boolean at1000) {
        in2 = (in2 & ~0x08) | (at1000 ? 0x08 : 0);
    }

    // For in2
    /** dip switch coin info: 0x00:on 0x80:off */
    public boolean getCoinInfoOff() {
        return (in2 & 0x80) != 0;
    }

    public void setCoinInfoOff(boolean coinInfo) {
        in2 = (in2 & ~0x80) | (coinInfo ? 0x80 : 0);
    }
}
