package atlasprobe;

/** Reuses the defined terrain but calls original DesertPyramidPiece.postProcess
 * and DesertPyramidStructure.afterPlace, including loot and archaeology. */
public final class StructurePyramidBlocksProbe {
    public static void main(String[] args) throws Exception {
        StructureHutBlocksProbe.run(true);
    }
}
