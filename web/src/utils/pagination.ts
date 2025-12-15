import { ref } from 'vue'

/**
 * A composable for handling pagination logic.
 * @param fetchData - The function to call when the page or page size changes.
 * @param initialPageSize - The initial page size.
 */
export function usePagination(fetchData: () => void, initialPageSize = 10) {
  const currentPage = ref(1)
  const pageSize = ref(initialPageSize)
  const total = ref(0)

  const handlePageChange = (page: number) => {
    currentPage.value = page
    fetchData()
  }

  const handleSizeChange = (size: number) => {
    pageSize.value = size
    currentPage.value = 1 // Reset to the first page
    fetchData()
  }

  // Function to reset pagination, e.g., when filters change
  const reset = () => {
    currentPage.value = 1
  }

  return {
    currentPage,
    pageSize,
    total,
    handlePageChange,
    handleSizeChange,
    reset
  }
}